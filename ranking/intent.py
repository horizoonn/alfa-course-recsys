import re
from dataclasses import dataclass, field

from text import normalize, tokens


@dataclass(frozen=True)
class Goal:
    name: str
    condition: str
    expansion: str
    subjects: tuple[str, ...]
    actions: tuple[str, ...]


GOALS = (
    Goal(
        "business_launch",
        r"(?:открыть|запустить|создать)\s+(?:(?:свой|свою|небольшую)\s+)?(?:бизнес|кофейню|кафе|магазин|салон)",
        "открыть бизнес запуск план",
        ("бизнес", "кофейн", "магазин", "кафе", "запуск"),
        ("откры", "запуск", "нача", "план"),
    ),
    Goal(
        "registration",
        r"(?:открыть|зарег[ие]стр\w*|регистрац\w*).{0,30}\b(?:ип|ооо)\b",
        "регистрация ип ооо документы открытие",
        ("ип", "ооо"),
        ("откры", "регистрац", "зарегистр"),
    ),
    Goal(
        "retention",
        r"(?:клиент|покупател).{0,65}(?:не\s+возвращ|один\s+раз)|(?:удержа|верну|повторн|лояльност).{0,40}(?:клиент|покуп|продаж)",
        "удержание клиентов повторные продажи лояльность покупателей",
        ("клиент", "покупател", "продаж"),
        ("удерж", "верну", "повторн", "лояльност"),
    ),
    Goal(
        "acquisition",
        r"(?:привле|найти|поиск|перв\w*|нет).{0,30}(?:клиент|покупател|заказчик)|(?:клиент|покупател).{0,20}(?:мало|нет)|мало.{0,20}(?:клиент|покупател)",
        "привлечение клиентов лидогенерация поиск покупателей",
        ("клиент", "покупател", "аудитор", "лид"),
        ("привлеч", "лидогенерац", "продвижен", "маркетинг"),
    ),
    Goal(
        "cashflow",
        r"(?:денег|деньги).{0,25}(?:не\s+хватает|нет)|(?:не\s+хватает|нет).{0,15}(?:денег|денежн)|денежн\w*\s+поток|кассов\w*\s+разрыв",
        "денежный поток cash flow кассовый разрыв",
        ("денежн", "кассов", "cash"),
        ("поток", "разрыв", "cash", "ддс"),
    ),
    Goal(
        "ad_efficiency",
        r"реклам.{0,35}(?:не\s+окуп|эффектив|рентабель)|эффектив\w*.{0,20}реклам",
        "эффективность рекламы окупаемость рекламных расходов",
        ("реклам",),
        ("эффективн", "окупаем", "результат", "отслед"),
    ),
    Goal(
        "sales",
        r"(?:увелич|повыс|рост).{0,25}продаж|продаж.{0,20}(?:падают|снижа)",
        "увеличить продажи конверсия покупателей",
        ("продаж", "конверс"),
        ("увелич", "повыс", "рост", "конверс"),
    ),
    Goal(
        "hiring",
        r"(?:наня|наним|подбор|найм).{0,25}(?:сотрудник|персонал|команд)|перв\w*\s+сотрудник",
        "найм подбор персонала первый сотрудник",
        ("сотрудник", "персонал", "найм", "подбор"),
        ("наня", "найм", "подбор", "приня"),
    ),
    Goal(
        "accounting",
        r"бухгалтер|(?:вести|ведение).{0,15}уч[её]т",
        "бухгалтерский учет отчетность",
        ("бухгалтер", "учет", "отчетност"),
        ("учет", "бухгалтер", "отчетност"),
    ),
    Goal(
        "export",
        r"экспорт|зарубежн\w*\s+рын",
        "экспорт внешнеэкономическая деятельность",
        ("экспорт", "внешнеэкономическ", "таможен"),
        ("экспорт", "таможен", "выход"),
    ),
)

CONTEXTS = {
    "food": r"кофейн|кафе\b|ресторан|общепит|пекарн",
    "retail": r"магазин|торговл|маркетплейс|розниц",
    "services": r"салон|парикмахер|услуг",
    "it": r"\bit\b|\bит\b|программ|разработк",
    "production": r"производств|завод|фабрик",
}
FORMATS = {
    "document": r"шаблон|образец|бланк|чек.?лист",
    "calculator": r"калькулятор",
    "article": r"стать[яюие]|статьи",
    "course": r"курс\w*",
    "podcast": r"подкаст\w*",
    "webinar": r"вебинар\w*",
    "term": r"определени\w*|термин\w*",
}


@dataclass
class Intent:
    original: str
    positive: str
    goals: list[Goal] = field(default_factory=list)
    contexts: list[str] = field(default_factory=list)
    free_only: bool = False
    excluded_kinds: list[str] = field(default_factory=list)
    allowed_kinds: list[str] = field(default_factory=list)
    excluded_terms: list[str] = field(default_factory=list)
    requested_kind: str | None = None
    definition: bool = False
    clarification: str = ""

    @property
    def search_text(self):
        return " ".join([self.positive, *(goal.expansion for goal in self.goals)])

    def diagnostic(self):
        return {
            "positive": self.positive,
            "search_text": self.search_text,
            "goals": [goal.name for goal in self.goals],
            "contexts": self.contexts,
            "excluded_terms": self.excluded_terms,
            "requested_kind": self.requested_kind,
            "definition": self.definition,
        }


def parse_intent(query):
    raw = normalize(query)
    intent = Intent(query, raw)
    free = r"только\s+бесплат\w*|без\s+платн\w*(?:\s+материал\w*)?|без\s+оплат\w*"
    intent.free_only = bool(re.search(free, raw))
    positive = re.sub(free, " ", raw)
    no_course = r"без\s+курс\w*|не\s+(?:нужны?|хочу|предлагайте)\s+(?:мне\s+)?курс\w*"
    if re.search(no_course, positive):
        intent.excluded_kinds.append("course")
        positive = re.sub(no_course, " ", positive)
    for term, pattern in {
        "франшиз": r"(?:без|не\s+хочу)\s+франшиз\w*",
        "скид": r"без\s+скид\w*",
    }.items():
        if re.search(pattern, positive):
            intent.excluded_terms.append(term)
            positive = re.sub(pattern, " ", positive)
    contrast = re.search(r"(?:\bа\s+не\s+|,\s*не\s+)(.+)$", positive)
    if contrast:
        excluded = contrast.group(1)
        intent.excluded_terms.extend(tokens(excluded))
        for kind, pattern in FORMATS.items():
            if re.search(pattern, excluded):
                intent.excluded_kinds.append(kind)
        positive = positive[: contrast.start()]
    for kind, pattern in FORMATS.items():
        if re.search(r"только\s+(?:коротк\w*\s+)?(?:" + pattern + ")", positive):
            intent.allowed_kinds.append(kind)
        if re.search(
            r"(?:нуж\w*|ищу|дайте|скачать|хочу|подберите)\s+(?:(?:готов\w*|коротк\w*)\s+)?(?:"
            + pattern
            + ")",
            positive,
        ):
            intent.requested_kind = kind
    if re.search(r"(?:составить|оформить|подготовить|скачать).{0,25}договор", positive):
        intent.requested_kind = "document"
    positive = re.sub(r"не\s+(?:знаю|понимаю)|без\s+(?:ошибок|проблем)", " ", positive)
    intent.positive = " ".join(positive.split()).strip(" ,.;!?")
    intent.definition = bool(
        re.search(r"что\s+такое|определени|значени\w*\s+термин", raw)
    )
    intent.contexts = [
        name
        for name, pattern in CONTEXTS.items()
        if re.search(pattern, intent.positive)
    ]
    intent.goals = [
        goal for goal in GOALS if re.search(goal.condition, intent.positive)
    ]
    if any(goal.name == "retention" for goal in intent.goals):
        intent.goals = [goal for goal in intent.goals if goal.name != "sales"]
    remaining = re.sub(r"не\s+(?:возвращ\w*|хватает|окуп\w*)", " ", intent.positive)
    if not tokens(intent.positive):
        intent.clarification = "Уточните, что хотите сделать: найти клиентов, спланировать платежи или подобрать документ."
    elif re.search(r"\b(?:не|без)\b", remaining):
        intent.clarification = "Уточните ограничение: какой результат нужен и какие способы решения исключить?"
    elif intent.allowed_kinds and all(
        kind in intent.excluded_kinds for kind in intent.allowed_kinds
    ):
        intent.clarification = "В запросе противоречат друг другу ограничения формата. Уточните, какой материал нужен."
    elif re.fullmatch(r"(?:нужен\s+)?договор", intent.positive):
        intent.clarification = (
            "Какой договор нужен: аренды, поставки, оказания услуг или трудовой?"
        )
    return intent
