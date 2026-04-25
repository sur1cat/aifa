"""
Обучение и сравнение моделей классификации расходов.
Запуск: python training/train.py

Модели:
  1. LogisticRegression  — мультиклассовая логрег (multinomial)
  2. LinearSVC           — метод опорных векторов (калиброванный)
  3. RandomForest        — ансамблевый метод на деревьях решений
  4. VotingEnsemble      — мягкое голосование LogReg + SVM
  5. SentenceTransformer — предобученные мультиязычные эмбеддинги + LogReg

Победитель по CV accuracy сохраняется как основная модель.
"""

import os
import sys
import json

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

import joblib
import numpy as np
from sklearn.calibration import CalibratedClassifierCV
from sklearn.ensemble import RandomForestClassifier, VotingClassifier
from sklearn.pipeline import Pipeline, FeatureUnion
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.linear_model import LogisticRegression
from sklearn.svm import LinearSVC
from sklearn.model_selection import cross_val_score, cross_val_predict
from sklearn.metrics import confusion_matrix, classification_report

from training.expense_data import TRAINING_DATA, CATEGORIES, CATEGORY_LABELS_RU, CATEGORY_LABELS_KZ
from app.encoder import MultilingualEncoder, ST_MODEL_NAME

MODELS_DIR = os.path.join(os.path.dirname(os.path.dirname(os.path.abspath(__file__))), "models")


# ── TF-IDF пайплайны ─────────────────────────────────────────────────────────

def _build_features() -> FeatureUnion:
    return FeatureUnion([
        ("char", TfidfVectorizer(
            analyzer="char_wb", ngram_range=(2, 4),
            max_features=8000, sublinear_tf=True, lowercase=True,
        )),
        ("word", TfidfVectorizer(
            analyzer="word", ngram_range=(1, 2),
            max_features=4000, sublinear_tf=True, lowercase=True,
        )),
    ])


def build_logreg() -> Pipeline:
    return Pipeline([
        ("features", _build_features()),
        ("clf", LogisticRegression(max_iter=2000, C=4.0,
                                   solver="lbfgs", multi_class="multinomial")),
    ])


def build_svm() -> Pipeline:
    return Pipeline([
        ("features", _build_features()),
        ("clf", CalibratedClassifierCV(LinearSVC(max_iter=3000, C=1.0), cv=3)),
    ])


def build_random_forest() -> Pipeline:
    return Pipeline([
        ("features", _build_features()),
        ("clf", RandomForestClassifier(n_estimators=200, max_depth=30,
                                       min_samples_leaf=2, n_jobs=-1, random_state=42)),
    ])


def build_voting_ensemble() -> VotingClassifier:
    return VotingClassifier(
        estimators=[("lr", build_logreg()), ("svm", build_svm())],
        voting="soft",
    )


def build_sentence_transformer() -> Pipeline:
    return Pipeline([
        ("encoder", MultilingualEncoder(ST_MODEL_NAME)),
        ("clf", LogisticRegression(max_iter=2000, C=4.0,
                                   solver="lbfgs", multi_class="multinomial")),
    ])


CANDIDATES = [
    ("LogisticRegression",   build_logreg),
    ("LinearSVC",            build_svm),
    ("RandomForest",         build_random_forest),
    ("VotingEnsemble",       build_voting_ensemble),
    ("SentenceTransformer",  build_sentence_transformer),
]


# ── Оценка ───────────────────────────────────────────────────────────────────

def _eval(builder_fn, texts: list, labels: list) -> tuple[float, float]:
    """CV accuracy для любого кандидата.

    Для SentenceTransformer эмбеддинги вычисляются один раз —
    это экономит время (не нужно загружать модель 5 раз).
    """
    estimator = builder_fn()

    if isinstance(estimator, Pipeline) and isinstance(
        estimator.named_steps.get("encoder"), MultilingualEncoder
    ):
        enc = MultilingualEncoder(ST_MODEL_NAME)
        enc.fit(texts)
        embeddings = enc.transform(texts)
        scores = cross_val_score(
            estimator.named_steps["clf"], embeddings, labels, cv=5, scoring="accuracy"
        )
    else:
        scores = cross_val_score(estimator, texts, labels, cv=5, scoring="accuracy")

    return float(scores.mean()), float(scores.std())


# ── Основная функция ──────────────────────────────────────────────────────────

def train():
    texts = [t for t, _ in TRAINING_DATA]
    labels = [l for _, l in TRAINING_DATA]

    print(f"Датасет: {len(texts)} примеров, {len(CATEGORIES)} категорий, 3 языка\n")
    print("── Сравнение моделей (5-fold CV) ──────────────────────────────")

    results: dict[str, dict] = {}
    for name, builder in CANDIDATES:
        mean, std = _eval(builder, texts, labels)
        results[name] = {"cv_accuracy": mean, "cv_std": std}
        print(f"  {name:<22} accuracy = {mean:.3f} ± {std:.3f}")

    winner_name = max(results, key=lambda n: results[n]["cv_accuracy"])
    winner_acc  = results[winner_name]["cv_accuracy"]
    winner_std  = results[winner_name]["cv_std"]

    print(f"\n  Победитель: {winner_name} ({winner_acc:.3f})")
    print("───────────────────────────────────────────────────────────────\n")

    # Обучаем победителя на полном датасете
    winner_builder = dict(CANDIDATES)[winner_name]
    pipeline = winner_builder()
    pipeline.fit(texts, labels)

    os.makedirs(MODELS_DIR, exist_ok=True)
    model_path = os.path.join(MODELS_DIR, "expense_classifier.joblib")
    joblib.dump(pipeline, model_path)
    print(f"Модель сохранена: {model_path}")

    # ── Детальные метрики победителя ─────────────────────────────────────
    print("\n── Детальные метрики победителя (cross_val_predict) ────────────")

    # Для ST — снова предвычисляем эмбеддинги
    if winner_name == "SentenceTransformer":
        enc = MultilingualEncoder(ST_MODEL_NAME)
        enc.fit(texts)
        embeddings = enc.transform(texts)
        cv_preds = cross_val_predict(
            dict(CANDIDATES)[winner_name]().named_steps["clf"],
            embeddings, labels, cv=5,
        )
    else:
        cv_preds = cross_val_predict(winner_builder(), texts, labels, cv=5)

    cm = confusion_matrix(labels, cv_preds, labels=CATEGORIES)
    report = classification_report(
        labels, cv_preds, labels=CATEGORIES,
        target_names=CATEGORIES, output_dict=True, zero_division=0,
    )

    print(f"  {'Категория':<18} {'Precision':>10} {'Recall':>8} {'F1':>8} {'Support':>9}")
    print("  " + "-" * 55)
    for cat in CATEGORIES:
        r = report[cat]
        print(f"  {cat:<18} {r['precision']:>10.3f} {r['recall']:>8.3f}"
              f" {r['f1-score']:>8.3f} {int(r['support']):>9}")
    print(f"\n  macro avg F1: {report['macro avg']['f1-score']:.3f}")
    print("───────────────────────────────────────────────────────────────\n")

    per_category = {
        cat: {
            "precision": round(report[cat]["precision"], 4),
            "recall":    round(report[cat]["recall"], 4),
            "f1":        round(report[cat]["f1-score"], 4),
            "support":   int(report[cat]["support"]),
            "label_ru":  CATEGORY_LABELS_RU[cat],
            "label_kz":  CATEGORY_LABELS_KZ[cat],
        }
        for cat in CATEGORIES
    }

    meta = {
        "categories":        CATEGORIES,
        "labels_ru":         CATEGORY_LABELS_RU,
        "labels_kz":         CATEGORY_LABELS_KZ,
        "n_training_samples": len(texts),
        "cv_accuracy":       winner_acc,
        "cv_std":            winner_std,
        "macro_f1":          round(report["macro avg"]["f1-score"], 4),
        "winner":            winner_name,
        "architecture":      f"FeatureUnion + {winner_name}" if winner_name != "SentenceTransformer"
                             else f"SentenceTransformer({ST_MODEL_NAME}) + LogisticRegression",
        "model_comparison":  results,
        "per_category":      per_category,
        "confusion_matrix":  {"labels": CATEGORIES, "matrix": cm.tolist()},
    }
    meta_path = os.path.join(MODELS_DIR, "expense_classifier_meta.json")
    with open(meta_path, "w", encoding="utf-8") as f:
        json.dump(meta, f, ensure_ascii=False, indent=2)
    print(f"Метаданные сохранены: {meta_path}\n")

    # ── Примеры предсказаний ──────────────────────────────────────────────
    print("── Примеры предсказаний ────────────────────────────────────────")
    samples = [
        ("Кофе латте",             "RU"),
        ("Яндекс Такси поездка",   "RU"),
        ("Ашан продукты",          "RU"),
        ("Wildberries заказ",      "RU"),
        ("Квартплата ЖКХ",         "RU"),
        ("Starbucks",              "EN"),
        ("Gym membership",         "EN"),
        ("Netflix",                "EN"),
        ("Amazon order",           "EN"),
        ("Авиабилет Москва Питер", "RU"),
        ("Мейрамхана",             "KZ"),
        ("Азық-түлік дүкені",      "KZ"),
        ("Автобус билеті",         "KZ"),
        ("Денсаулық клиникасы",    "KZ"),
        ("Kaspi аударым",          "KZ"),
    ]
    for s, lang in samples:
        proba = pipeline.predict_proba([s])[0]
        top_idx = np.argmax(proba)
        cat  = pipeline.classes_[top_idx]
        conf = proba[top_idx]
        label_ru = CATEGORY_LABELS_RU[cat]
        label_kz = CATEGORY_LABELS_KZ[cat]
        mark = "✓" if conf >= 0.55 else "?"
        print(f"  [{lang}] {mark} '{s}' → {cat} | RU: {label_ru} | KZ: {label_kz} | conf={conf:.2f}")


if __name__ == "__main__":
    train()
