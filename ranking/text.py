import re
from functools import lru_cache

from nltk.stem.snowball import RussianStemmer

GROUPS = {
    "title": 0,
    "summary": 0,
    "metadata": 0,
    "body": 1,
    "caption": 2,
    "reference": 2,
}
WEIGHTS = (0.7, 0.3, 0.05)
WORDS = re.compile(r"[^\W_]+", re.UNICODE)
STEMMER = RussianStemmer()
SHORT = {"ип", "ии", "hr", "it"}
STOP = frozenset(
    STEMMER.stem(word)
    for word in (
        "как что для или это при его она они вам вас все уже наш свой свое "
        "бизнес предприниматель материалы формат сфера "
        "хочу хотеть нужно нужен нужна помогите помощь разобраться понять научиться "
        "могу можно подскажите пожалуйста получить платформа"
    ).split()
)


@lru_cache(maxsize=65536)
def stem(word):
    return STEMMER.stem(word)


def tokens(text):
    return [
        value
        for word in WORDS.findall(text.casefold().replace("ё", "е"))
        if (len(word) >= 3 or word in SHORT) and (value := stem(word)) not in STOP
    ]


def normalize(text):
    return " ".join(text.casefold().replace("ё", "е").split())
