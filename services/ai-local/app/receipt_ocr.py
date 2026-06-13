import os
import re
from concurrent.futures import ThreadPoolExecutor
from dataclasses import dataclass
from datetime import datetime
from difflib import SequenceMatcher
from io import BytesIO
from functools import lru_cache
from typing import Optional

from PIL import Image, ImageFilter, ImageOps

from app.model import get_classifier


TOTAL_HINTS = (
    "итого",
    "итог",
    "сумма",
    "к оплате",
    "барлығы",
    "жалпы",
    "total",
    "amount due",
)

SECTION_BREAK_HINTS = (
    "сатылым",
    "продажа",
    "продама",
    "итого",
    "барлығы",
    "кассир",
    "оператор",
    "документ",
    "проверка чека",
    "consumer.",
    "офд",
    "фиск.",
    "спасибо",
    "товара",
)

LEGAL_FORM_HINTS = (
    "жауапкершілігі",
    "шектеулі серіктестігі",
    "товарищество",
    "ответственностью",
    "тоо",
    "тoo",
    "llp",
    "ип",
    "ао ",
    "ao ",
)

HEADER_NOISE_HINTS = (
    "кассовый чек",
    "cash receipt",
    "место расчетов",
    "павильон",
    "адрес",
    "чек на продажу",
    "чек на приход",
    "кассовый",
)

FOOTER_HINTS = (
    "документ",
    "смена",
    "кассир",
    "инн",
    "рн ккт",
    "фн",
    "фд",
    "фп",
)

CYRILLIC_LOOKALIKE_TO_LATIN = str.maketrans(
    {
        "А": "A",
        "В": "B",
        "Е": "E",
        "К": "K",
        "М": "M",
        "Н": "H",
        "О": "O",
        "Р": "P",
        "С": "C",
        "Т": "T",
        "У": "Y",
        "Х": "X",
        "а": "a",
        "е": "e",
        "о": "o",
        "р": "p",
        "с": "c",
        "у": "y",
        "х": "x",
        "к": "k",
        "м": "m",
        "т": "t",
        "в": "b",
    }
)

# Reverse map: Latin OCR-lookalikes → Cyrillic, for hint matching on mixed OCR output.
# Tesseract sometimes reads Cyrillic letters as their Latin look-alikes (e.g. "CYMMA" instead of "СУММА").
LATIN_LOOKALIKE_TO_CYRILLIC = str.maketrans(
    {
        "A": "А", "B": "В", "E": "Е", "K": "К", "M": "М", "H": "Н",
        "O": "О", "P": "Р", "C": "С", "T": "Т", "Y": "У", "X": "Х",
        "a": "а", "e": "е", "o": "о", "p": "р", "c": "с", "y": "у",
        "x": "х", "k": "к", "m": "м", "t": "т", "b": "в",
    }
)

# Pattern to detect document/receipt serial numbers or tax IDs that should not be treated as amounts.
# e.g. "ЧЕК НА ПРОДАЖУ N:277087", "N:12345", "№277087", "ИИН 600464", "БИН 123456"
_DOC_NUMBER_RE = re.compile(
    r"(?:чек\s+на\s+(?:продажу|приход)"
    r"|n\s*:\s*\d{4,}"
    r"|№\s*\d{4,}"
    r"|#\s*\d{4,}"
    r"|(?:иин|iin|бин|bin|инн|inn|рнн|кпп|бсн|рн\s+ккт|рн\s*ккт)\s*\d{4,}"
    r"|эвд\s*n\s*:\s*\d+"
    r")",
    re.IGNORECASE,
)

DATE_PATTERNS = [
    re.compile(r"\b(\d{2}[./-]\d{2}[./-]\d{4})(?:[,\s]+(\d{2}:\d{2}(?::\d{2})?))?\b"),
    re.compile(r"\b(\d{4}[./-]\d{2}[./-]\d{2})(?:[,\s]+(\d{2}:\d{2}(?::\d{2})?))?\b"),
]
DATETIME_PATTERNS = [
    re.compile(r"\b(\d{2}[./-]\d{2}[./-]\d{4})[,\s]+(\d{2}:\d{2}(?::\d{2})?)\b"),
    re.compile(r"\b(\d{4}[./-]\d{2}[./-]\d{2})[,\s]+(\d{2}:\d{2}(?::\d{2})?)\b"),
]

# Текстовые названия месяцев (RU / EN) → номер месяца
_MONTHS_RU: dict[str, int] = {
    "январ": 1, "феврал": 2, "март": 3, "апрел": 4,
    "мая": 5, "май": 5, "июн": 6, "июл": 7,
    "август": 8, "сентябр": 9, "октябр": 10, "ноябр": 11, "декабр": 12,
}
_MONTHS_EN: dict[str, int] = {
    "jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5,
    "jun": 6, "jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

# DD <месяц-RU/EN> YYYY [HH:MM[:SS]]
_TEXT_DATE_DMY = re.compile(
    r"\b(\d{1,2})\s+([а-яёa-z]+)\s+(\d{4})(?:[,\s]+(\d{2}:\d{2}(?::\d{2})?))?",
    re.IGNORECASE | re.UNICODE,
)
# <месяц-EN> DD,? YYYY [HH:MM[:SS]]
_TEXT_DATE_MDY = re.compile(
    r"\b([a-z]+)\s+(\d{1,2}),?\s+(\d{4})(?:[,\s]+(\d{2}:\d{2}(?::\d{2})?))?",
    re.IGNORECASE,
)

MONEY_RE = re.compile(r"(?<!\d)(\d{1,3}(?:[ \u00a0']?\d{3})*(?:[.,]\d{1,2})?|\d+(?:[.,]\d{1,2})?)(?!\d)")
TIME_RE = re.compile(r"\b(\d{2}:\d{2}(?::\d{2})?)\b")


@dataclass
class ReceiptScanResult:
    amount: Optional[float]
    currency: str
    date: Optional[str]
    merchant: str
    category: str
    label_ru: str
    label_kz: str
    items: list[str]
    confidence: float
    raw_total: str
    raw_text: str


def _parse_money(raw: str) -> Optional[float]:
    candidate = raw.replace("\u00a0", " ").replace(" ", "").replace(",", ".").replace("'", "")
    try:
        value = float(candidate)
    except ValueError:
        return None
    return value if value > 0 else None


def _detect_currency(text: str) -> str:
    lower = text.lower()
    if any(token in lower for token in ("₸", "тенге", "тг", "kzt")):
        return "KZT"
    if any(token in lower for token in ("руб", "₽", "rub")):
        return "RUB"
    if any(token in lower for token in ("$", "usd")):
        return "USD"
    if any(token in lower for token in ("€", "eur")):
        return "EUR"
    return "KZT"


def _normalize_lines(text: str) -> list[str]:
    lines = []
    for raw in text.splitlines():
        line = " ".join(raw.strip().split())
        if line:
            lines.append(line)
    return lines


def _score_ocr_text(text: str) -> int:
    score = 0
    lines = _normalize_lines(text)
    if len(lines) >= 6:
        score += 2
    lower = text.lower()
    if any(hint in lower for hint in TOTAL_HINTS):
        score += 3
    if any(pattern.search(text) for pattern in DATE_PATTERNS):
        score += 2
    if any(token in lower for token in ("₸", "тенге", "тг", "kzt")):
        score += 1
    score += min(4, len(text.strip()) // 120)
    return score


def _contains_hint(line: str, hints: tuple[str, ...]) -> bool:
    # Check both the raw lowercased line and a version where Latin OCR lookalikes
    # are converted to Cyrillic (e.g. "CYMMA" → "СУММА").
    lowered = line.lower()
    normalized = line.translate(LATIN_LOOKALIKE_TO_CYRILLIC).lower()
    for hint in hints:
        pattern = rf"(?<![a-zа-я]){re.escape(hint)}(?![a-zа-я])"
        if re.search(pattern, lowered) or re.search(pattern, normalized):
            return True
    return False


def _preprocess_variants(image: Image.Image) -> list[Image.Image]:
    base = ImageOps.exif_transpose(image).convert("RGB")
    max_width = 1600
    if base.width > max_width:
        ratio = max_width / base.width
        base = base.resize((max_width, int(base.height * ratio)), Image.LANCZOS)
    elif base.width < 800:
        scale = max(2, int(1600 / max(base.width, 1)))
        base = base.resize((base.width * scale, base.height * scale))

    grayscale = ImageOps.grayscale(base)
    autocontrast = ImageOps.autocontrast(grayscale)
    sharpened = autocontrast.filter(ImageFilter.SHARPEN)
    threshold = autocontrast.point(lambda p: 255 if p > 168 else 0)
    return [autocontrast, sharpened, threshold]


def _ocr_image(image: Image.Image, languages: str, psm_modes: tuple[str, ...]) -> list[str]:
    import pytesseract

    variants = _preprocess_variants(image)
    tasks = [(variant, psm) for variant in variants for psm in psm_modes]

    def _run_one(args: tuple) -> str:
        variant, psm = args
        try:
            return pytesseract.image_to_string(variant, lang=languages, config=f"--oem 3 --psm {psm}")
        except Exception:
            try:
                return pytesseract.image_to_string(variant, lang="eng+rus", config=f"--oem 3 --psm {psm}")
            except Exception:
                return ""

    with ThreadPoolExecutor(max_workers=min(len(tasks), 2)) as pool:
        return [text for text in pool.map(_run_one, tasks) if text.strip()]


def _load_image(image_bytes: bytes) -> Image.Image:
    return ImageOps.exif_transpose(Image.open(BytesIO(image_bytes))).convert("RGB")


def _extract_text(image: Image.Image, languages: str) -> str:
    try:
        import pytesseract  # noqa: F401  ensure installed
    except ImportError as exc:
        raise RuntimeError("pytesseract is not installed") from exc

    candidates = _ocr_image(image, languages, ("4", "6", "11"))
    return max(candidates, key=_score_ocr_text) if candidates else ""


def _crop_zone(image: Image.Image, top_ratio: float, bottom_ratio: float) -> Image.Image:
    top = int(image.height * top_ratio)
    bottom = int(image.height * bottom_ratio)
    return image.crop((0, top, image.width, max(top + 1, bottom)))


def _extract_zone_candidates(
    image: Image.Image,
    languages: str,
    *,
    top_ratio: float,
    bottom_ratio: float,
    psm_modes: tuple[str, ...],
) -> list[str]:
    zone = _crop_zone(image, top_ratio, bottom_ratio)
    return _ocr_image(zone, languages, psm_modes)


def _extract_header_candidates(image: Image.Image, languages: str) -> list[str]:
    return _extract_zone_candidates(
        image,
        languages,
        top_ratio=0.0,
        bottom_ratio=0.22,
        psm_modes=("6", "7", "11", "13"),
    )


def _extract_footer_candidates(image: Image.Image, languages: str) -> list[str]:
    return _extract_zone_candidates(
        image,
        languages,
        top_ratio=0.68,
        bottom_ratio=1.0,
        psm_modes=("4", "6", "11"),
    )


def _extract_total_candidates(image: Image.Image, languages: str) -> list[str]:
    return _extract_zone_candidates(
        image,
        languages,
        top_ratio=0.5,
        bottom_ratio=0.82,
        psm_modes=("4", "6", "11"),
    )


def _extract_date(lines: list[str], footer_lines: Optional[list[str]] = None) -> Optional[str]:
    search_lines = list(lines)
    if footer_lines:
        search_lines = [*footer_lines, *search_lines]

    stitched_lines = [
        f"{search_lines[idx]} {search_lines[idx + 1]}"
        for idx in range(len(search_lines) - 1)
    ]
    stitched_lines.extend(search_lines)

    for line in stitched_lines:
        for pattern in DATETIME_PATTERNS:
            match = pattern.search(line)
            if match:
                raw_date = match.group(1).replace("/", "-").replace(".", "-")
                raw_time = match.group(2)
                parsed: Optional[datetime] = None
                for fmt in ("%d-%m-%Y", "%Y-%m-%d"):
                    try:
                        parsed = datetime.strptime(raw_date, fmt)
                        break
                    except ValueError:
                        continue
                if parsed is None:
                    continue
                try:
                    time_value = datetime.strptime(raw_time, "%H:%M:%S")
                except ValueError:
                    try:
                        time_value = datetime.strptime(raw_time, "%H:%M")
                    except ValueError:
                        time_value = None
                if time_value is None:
                    continue
                parsed = parsed.replace(
                    hour=time_value.hour,
                    minute=time_value.minute,
                    second=time_value.second,
                )
                return parsed.isoformat(timespec="seconds")

    for line in stitched_lines:
        for pattern in DATE_PATTERNS:
            match = pattern.search(line)
            if match:
                raw_date = match.group(1).replace("/", "-").replace(".", "-")
                raw_time = match.group(2)
                parsed: Optional[datetime] = None
                for fmt in ("%d-%m-%Y", "%Y-%m-%d"):
                    try:
                        parsed = datetime.strptime(raw_date, fmt)
                        break
                    except ValueError:
                        continue
                if parsed is None:
                    continue
                if raw_time:
                    try:
                        time_value = datetime.strptime(raw_time, "%H:%M:%S")
                    except ValueError:
                        try:
                            time_value = datetime.strptime(raw_time, "%H:%M")
                        except ValueError:
                            time_value = None
                    if time_value is not None:
                        parsed = parsed.replace(
                            hour=time_value.hour,
                            minute=time_value.minute,
                            second=time_value.second,
                        )
                        return parsed.isoformat(timespec="seconds")
                return parsed.date().isoformat()

    # ── Третий проход: текстовые названия месяцев (RU / EN) ──────────────────
    for line in stitched_lines:
        result = _parse_text_month_date(line)
        if result:
            return result
    return None


def _resolve_month(word: str) -> Optional[int]:
    """'марта' / 'March' / 'mar' → номер месяца или None."""
    lower = word.lower()
    for prefix, num in _MONTHS_RU.items():
        if lower.startswith(prefix):
            return num
    for prefix, num in _MONTHS_EN.items():
        if lower.startswith(prefix):
            return num
    return None


def _apply_time(parsed: datetime, raw_time: Optional[str]) -> datetime:
    if not raw_time:
        return parsed
    for fmt in ("%H:%M:%S", "%H:%M"):
        try:
            t = datetime.strptime(raw_time, fmt)
            return parsed.replace(hour=t.hour, minute=t.minute, second=t.second)
        except ValueError:
            continue
    return parsed


def _parse_text_month_date(line: str) -> Optional[str]:
    """Извлекает дату вида '15 марта 2025 10:30' или 'March 15, 2025'."""
    # DD MONTHNAME YYYY
    m = _TEXT_DATE_DMY.search(line)
    if m:
        day, month_word, year, raw_time = int(m.group(1)), m.group(2), int(m.group(3)), m.group(4)
        month = _resolve_month(month_word)
        if month and 1 <= day <= 31 and 1900 <= year <= 2100:
            try:
                parsed = _apply_time(datetime(year, month, day), raw_time)
                return parsed.isoformat(timespec="seconds") if raw_time else parsed.date().isoformat()
            except ValueError:
                pass

    # MONTHNAME DD,? YYYY
    m = _TEXT_DATE_MDY.search(line)
    if m:
        month_word, day, year, raw_time = m.group(1), int(m.group(2)), int(m.group(3)), m.group(4)
        month = _resolve_month(month_word)
        if month and 1 <= day <= 31 and 1900 <= year <= 2100:
            try:
                parsed = _apply_time(datetime(year, month, day), raw_time)
                return parsed.isoformat(timespec="seconds") if raw_time else parsed.date().isoformat()
            except ValueError:
                pass
    return None


def _extract_total(lines: list[str]) -> tuple[Optional[float], str]:
    discounted: list[tuple[float, str, int]] = []
    pre_discount: list[tuple[float, str, int]] = []
    prioritized: list[tuple[float, str, int]] = []
    fallback: list[tuple[float, str, int]] = []

    for idx, line in enumerate(lines):
        # Skip document/receipt number lines — their serial numbers look like amounts.
        if _DOC_NUMBER_RE.search(line):
            continue
        matches = MONEY_RE.findall(line)
        if not matches:
            continue
        parsed = [(_parse_money(raw), raw) for raw in matches]
        parsed = [(value, raw) for value, raw in parsed if value is not None]
        if not parsed:
            continue

        line_lower = line.lower()
        for value, raw in parsed:
            if TIME_RE.search(line):
                continue
            if value < 1:
                continue
            if any(hint in line_lower for hint in _DISCOUNT_APPLIED_HINTS):
                discounted.append((value, raw, idx))
            elif any(hint in line_lower for hint in _PRE_DISCOUNT_HINTS):
                pre_discount.append((value, raw, idx))
            elif _contains_hint(line_lower, TOTAL_HINTS):
                prioritized.append((value, raw, idx))
            else:
                fallback.append((value, raw, idx))

    # Priority: after-discount > regular total hints > pre-discount > fallback
    candidates = discounted or prioritized or pre_discount or fallback
    if not candidates:
        return None, ""

    # Filter obviously wrong amounts (fiscal doc numbers, barcodes)
    # Real receipts in KZT are typically 1 - 5_000_000
    realistic = [(v, r, i) for v, r, i in candidates if 1 <= v <= 5_000_000]
    if not realistic:
        realistic = [(v, r, i) for v, r, i in candidates if v <= 100_000_000]
    if not realistic:
        realistic = candidates

    # If we are relying only on fallback (no total-hint line found), reject large
    # whole integers (>= 50 000) without a fractional part — these are almost always
    # document/fiscal serial numbers, not purchase amounts.
    using_only_fallback = not (discounted or prioritized or pre_discount)
    if using_only_fallback:
        filtered = [(v, r, i) for v, r, i in realistic if not (v >= 50_000 and v == int(v))]
        if filtered:
            realistic = filtered

    amount, raw, _ = max(realistic, key=lambda item: item[0])
    return amount, raw


_DISCOUNT_APPLIED_HINTS = ("с учетом скидок", "с учётом скидок", "после скидок", "с учетом скидки", "с учётом скидки")
_PRE_DISCOUNT_HINTS = ("без скидок", "до скидок")


def _extract_total_from_zone(lines: list[str]) -> tuple[Optional[float], str]:
    discounted: tuple[Optional[float], str] = (None, "")
    pre_discount: tuple[Optional[float], str] = (None, "")
    regular: tuple[Optional[float], str] = (None, "")

    for line in lines:
        if _DOC_NUMBER_RE.search(line):
            continue
        lowered = line.lower()
        if not _contains_hint(lowered, TOTAL_HINTS):
            continue
        matches = MONEY_RE.findall(line)
        parsed = [(_parse_money(raw), raw) for raw in matches]
        parsed = [(value, raw) for value, raw in parsed if value is not None]
        if not parsed:
            continue
        amount, raw = max(parsed, key=lambda item: item[0])
        if amount is None:
            continue

        if any(hint in lowered for hint in _DISCOUNT_APPLIED_HINTS):
            if discounted[0] is None:
                discounted = (amount, raw)
        elif any(hint in lowered for hint in _PRE_DISCOUNT_HINTS):
            if pre_discount[0] is None:
                pre_discount = (amount, raw)
        else:
            if regular[0] is None:
                regular = (amount, raw)

    # Priority: after-discount total > regular total > pre-discount total
    return discounted if discounted[0] else (regular if regular[0] else pre_discount)


def _clean_merchant_line(line: str) -> str:
    import re as _re2
    line = _re2.sub(r"^(ООО|ТОО|ИП|АО|ЗАО|ОАО|ПАО|LLP|LLC)\s*", "", line.strip(), flags=_re2.IGNORECASE).strip()
    line = line.strip().strip('"').strip("'").strip("`").strip("“”«»")
    return re.sub(r"\s+", " ", line)[:80]


def _strip_merchant_suffixes(value: str) -> str:
    cleaned = _clean_merchant_line(value)
    for suffix in LEGAL_FORM_HINTS:
        pos = cleaned.lower().find(suffix)
        if pos > 0:
            cleaned = cleaned[:pos].strip(" ,.-")
            break
    return cleaned


def _merchant_match_key(value: str) -> str:
    normalized = _strip_merchant_suffixes(value).translate(CYRILLIC_LOOKALIKE_TO_LATIN)
    normalized = normalized.lower()
    normalized = re.sub(r"[^a-z0-9]+", "", normalized)
    return normalized


def _merchant_latin_score(value: str) -> int:
    translated = value.translate(CYRILLIC_LOOKALIKE_TO_LATIN)
    return sum("A" <= ch <= "Z" or "a" <= ch <= "z" for ch in translated)


@lru_cache(maxsize=1)
def _merchant_aliases() -> dict[str, str]:
    raw = os.getenv("OCR_MERCHANT_ALIASES", "")
    aliases: dict[str, str] = {}
    for chunk in raw.split(";"):
        chunk = chunk.strip()
        if not chunk or "=" not in chunk:
            continue
        source, target = chunk.split("=", 1)
        target = _strip_merchant_suffixes(target)
        key = _merchant_match_key(source)
        if key and target:
            aliases[key] = target
    return aliases


@lru_cache(maxsize=1)
def _merchant_hints() -> tuple[str, ...]:
    raw = os.getenv("OCR_MERCHANT_HINTS", "")
    hints = [
        _strip_merchant_suffixes(item)
        for item in raw.split(";")
        for item in item.split(",")
    ]
    return tuple(item for item in hints if item)


def _correct_merchant_with_hints(candidate: str) -> str:
    aliases = _merchant_aliases()
    candidate_key = _merchant_match_key(candidate)
    if candidate_key in aliases:
        return aliases[candidate_key]

    hints = _merchant_hints()
    if not hints:
        return candidate

    best_hint = candidate
    best_ratio = 0.0
    candidate_suffix = candidate.split()[-1].lower() if candidate.split() else ""
    for hint in hints:
        hint_suffix = hint.split()[-1].lower() if hint.split() else ""
        suffix_bonus = 0.04 if candidate_suffix and candidate_suffix == hint_suffix else 0.0
        ratio = SequenceMatcher(
            None,
            _merchant_match_key(candidate),
            _merchant_match_key(hint),
        ).ratio() + suffix_bonus
        if ratio > best_ratio:
            best_ratio = ratio
            best_hint = hint

    return best_hint if best_ratio >= 0.84 else candidate


def _merchant_score(line: str, index: int) -> int:
    lowered = line.lower()
    if any(hint in lowered for hint in SECTION_BREAK_HINTS):
        return -100
    if any(hint in lowered for hint in HEADER_NOISE_HINTS):
        return -100
    if any(token in lowered for token in ("http://", "https://", "consumer.", "офд")):
        return -100
    if len(line) < 3:
        return -100
    if sum(ch.isdigit() for ch in line) > max(2, len(line) // 4):
        return -40

    score = 0
    if index < 4:
        score += 6
    elif index < 8:
        score += 3
    if '"' in line or "«" in line or "“" in line:
        score += 4
    if any("A" <= ch <= "Z" or "a" <= ch <= "z" for ch in line):
        score += 3
    if any("А" <= ch <= "Я" or "а" <= ch <= "я" for ch in line):
        score += 2
    if 4 <= len(line) <= 42:
        score += 3
    if any(hint in lowered for hint in LEGAL_FORM_HINTS):
        if chr(171) in line or chr(8220) in line or '"' in line:
            score += 5
        else:
            score -= 3
    if any(
        token in lowered
        for token in (
            "г. ",
            "ул.",
            "street",
            "st.",
            "road",
            "avenue",
            "ave",
            "blvd",
            "жсн",
            "бин",
            "иин",
            "документ",
            "document",
            "оператор",
        )
    ):
        score -= 8
    if any(ch.isdigit() for ch in line):
        score -= 5
    return score


# Common OCR-misread or generic receipt words that are never valid merchant names
_INVALID_MERCHANT_WORDS: frozenset[str] = frozenset({
    "чек", "чак", "час", "чак", "наличные", "наличными", "итого", "сумма",
    "receipt", "total", "cash", "change",
})


def _is_valid_merchant(name: str) -> bool:
    """Return False for obvious OCR garbage (too many non-alpha chars, punctuation typical of address/code lines)."""
    if len(name) < 3:
        return False
    # Exact match against known non-merchant words
    if name.lower().strip() in _INVALID_MERCHANT_WORDS:
        return False
    if any(ch in name for ch in (':', '/', '%', '@', '#', '|', '\\', '=', '*', ')', '(')):
        return False
    alpha = sum(ch.isalpha() for ch in name)
    if alpha < len(name) * 0.45:
        return False
    digits = sum(ch.isdigit() for ch in name)
    if digits > len(name) * 0.3:
        return False
    return True


def _extract_merchant(lines: list[str], header_lines: list[str]) -> str:
    grouped_candidates: dict[str, tuple[int, int, str]] = {}
    combined = header_lines + lines[:12]
    for idx, line in enumerate(combined):
        score = _merchant_score(line, idx)
        if score <= 0:
            continue
        extracted = None
        quoted = re.search(r"[\"«“]([^\"»”]+)[\"»”]", line)
        if quoted:
            extracted = quoted.group(1)
        else:
            extracted = line

        cleaned = _strip_merchant_suffixes(extracted)
        if len(cleaned) < 3:
            continue

        key = _merchant_match_key(cleaned)
        if not key:
            continue

        latin_score = _merchant_latin_score(cleaned)
        existing = grouped_candidates.get(key)
        aggregate_score = score + latin_score
        if existing is None:
            grouped_candidates[key] = (aggregate_score, latin_score, cleaned)
            continue

        total_score, best_latin_score, best_value = existing
        total_score += aggregate_score
        if latin_score > best_latin_score or (
            latin_score == best_latin_score and len(cleaned) > len(best_value)
        ):
            best_value = cleaned
            best_latin_score = latin_score
        grouped_candidates[key] = (total_score, best_latin_score, best_value)

    if not grouped_candidates:
        return ""

    best_candidate = max(
        grouped_candidates.values(),
        key=lambda item: (item[0], item[1], len(item[2])),
    )[2]

    if not _is_valid_merchant(best_candidate):
        return ""

    return _correct_merchant_with_hints(best_candidate)


def _extract_items(lines: list[str], merchant: str, raw_total: str, receipt_date: Optional[str]) -> list[str]:
    items: list[str] = []
    has_explicit_items_block = any(
        any(hint in line.lower() for hint in ("сатылым", "продажа", "товар", "позиция", "sales"))
        for line in lines
    )
    in_items_block = False
    idx = 0
    while idx < len(lines):
        line = lines[idx]
        lowered = line.lower()
        if merchant and line == merchant:
            idx += 1
            continue
        if receipt_date and receipt_date[:10] in line:
            idx += 1
            continue
        if any(hint in lowered for hint in ("сатылым", "продажа", "товар", "позиция", "sales")):
            in_items_block = True
            idx += 1
            continue
        if raw_total and raw_total in line and any(hint in lowered for hint in TOTAL_HINTS):
            break
        if any(hint in lowered for hint in TOTAL_HINTS):
            break
        if any(token in lowered for token in FOOTER_HINTS):
            idx += 1
            continue
        if any(token in lowered for token in ("http://", "https://", "consumer.", "офд", "рекламный")):
            idx += 1
            continue
        if len(line) < 3:
            idx += 1
            continue
        alpha_count = sum(ch.isalpha() for ch in line)
        if alpha_count < 2:
            idx += 1
            continue
        if has_explicit_items_block and not in_items_block:
            idx += 1
            continue
        if not in_items_block and any(
            token in lowered
            for token in (
                *HEADER_NOISE_HINTS,
                *LEGAL_FORM_HINTS,
                "ооо",
                "г.",
                "ул.",
                "дом",
                "адрес",
                "место расчетов",
                "павильон",
            )
        ):
            idx += 1
            continue
        has_money = bool(MONEY_RE.search(line))
        next_line = lines[idx + 1] if idx + 1 < len(lines) else ""
        next_has_money = bool(MONEY_RE.search(next_line))
        looks_like_item_name_with_size = (
            alpha_count >= 4
            and next_has_money
            and "*" not in line
            and "=" not in line
            and not any(token in lowered for token in TOTAL_HINTS + FOOTER_HINTS)
        )
        if looks_like_item_name_with_size:
            in_items_block = True
            items.append(line[:80])
            idx += 2
            if len(items) == 8:
                break
            continue
        if not has_money and alpha_count >= 3 and next_has_money:
            in_items_block = True
            items.append(line[:80])
            idx += 2
            if len(items) == 8:
                break
            continue
        if in_items_block and not has_money and alpha_count >= 3 and not next_has_money:
            items.append(line[:80])
            idx += 1
            if len(items) == 8:
                break
            continue
        idx += 1
    return items


def scan_receipt(image_bytes: bytes, languages: Optional[str] = None) -> ReceiptScanResult:
    if languages is None:
        languages = os.getenv("OCR_LANGS", "eng+rus+kaz")

    image = _load_image(image_bytes)

    with ThreadPoolExecutor(max_workers=2) as pool:
        fut_text = pool.submit(_extract_text, image, languages)
        fut_header = pool.submit(_extract_header_candidates, image, languages)
        fut_footer = pool.submit(_extract_footer_candidates, image, languages)
        fut_total = pool.submit(_extract_total_candidates, image, languages)
        raw_text = fut_text.result()
        header_candidates = fut_header.result()
        footer_candidates = fut_footer.result()
        total_candidates = fut_total.result()

    lines = _normalize_lines(raw_text)
    header_lines: list[str] = []
    for text in header_candidates:
        header_lines.extend(_normalize_lines(text))
    footer_lines: list[str] = []
    for text in footer_candidates:
        footer_lines.extend(_normalize_lines(text))
    total_lines: list[str] = []
    for text in total_candidates:
        total_lines.extend(_normalize_lines(text))

    amount, raw_total = _extract_total_from_zone(total_lines)
    if amount is None:
        amount, raw_total = _extract_total(lines)
    receipt_date = _extract_date(lines, footer_lines=footer_lines)
    merchant = _extract_merchant(lines, header_lines)
    items = _extract_items(lines, merchant, raw_total, receipt_date)
    currency = _detect_currency(raw_text)

    category = "shopping"
    label_ru = "Покупки"
    label_kz = "Сатып алулар"
    category_confidence = 0.0

    classify_text = " ".join(filter(None, [merchant, *items]))
    if classify_text.strip():
        result = get_classifier().predict(classify_text)
        category = result.category
        label_ru = result.label_ru
        label_kz = result.label_kz
        category_confidence = result.confidence

    confidence = 0.2
    if raw_text.strip():
        confidence += 0.2
    if merchant:
        confidence += 0.15
    if amount is not None:
        confidence += 0.2
    if receipt_date:
        confidence += 0.1
    confidence += min(0.15, category_confidence * 0.15)
    confidence = round(min(confidence, 0.95), 4)

    return ReceiptScanResult(
        amount=amount,
        currency=currency,
        date=receipt_date,
        merchant=merchant,
        category=category,
        label_ru=label_ru,
        label_kz=label_kz,
        items=items,
        confidence=confidence,
        raw_total=raw_total,
        raw_text=raw_text,
    )
