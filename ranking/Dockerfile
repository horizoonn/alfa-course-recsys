FROM ghcr.io/astral-sh/uv:0.11.21 AS uv
FROM python:3.12-slim-bookworm
COPY --from=uv /uv /usr/local/bin/uv
WORKDIR /app
COPY pyproject.toml uv.lock ./
RUN uv sync --frozen --no-dev --no-editable
COPY prepare_text.py prepare_search.py search_index.py search_api.py semantic.py text.py intent.py ./
USER 1000:1000
ENTRYPOINT ["/app/.venv/bin/python"]
CMD ["prepare_search.py", "/data/ranking_catalog.json", "--output", "/data/ranking_text_scores.json", "--search-dir", "/data/search"]
