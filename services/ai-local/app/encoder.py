from sklearn.base import BaseEstimator, TransformerMixin

ST_MODEL_NAME = "paraphrase-multilingual-MiniLM-L12-v2"


class MultilingualEncoder(BaseEstimator, TransformerMixin):
    def __init__(self, model_name: str = ST_MODEL_NAME):
        self.model_name = model_name
        self._encoder = None

    def _load(self):
        if self._encoder is None:
            from sentence_transformers import SentenceTransformer
            self._encoder = SentenceTransformer(self.model_name)
        return self._encoder

    def fit(self, X, y=None):
        self._load()
        return self

    def transform(self, X):
        return self._load().encode(list(X), show_progress_bar=False)

    def __getstate__(self):
        state = self.__dict__.copy()
        state["_encoder"] = None
        return state
