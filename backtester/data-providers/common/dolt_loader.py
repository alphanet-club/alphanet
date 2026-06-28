"""Small Dolt helpers shared by provider importer scripts."""

from __future__ import annotations

import datetime as dt
import json
import subprocess
from pathlib import Path
from typing import Any, Iterable, Sequence


class DoltError(RuntimeError):
    """Raised when a Dolt command fails."""


def utc_now() -> str:
    return dt.datetime.now(dt.timezone.utc).replace(microsecond=0).strftime("%Y-%m-%d %H:%M:%S")


def run_dolt(db_path: Path, args: Sequence[str], *, check: bool = True) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(
        ["dolt", *args],
        cwd=db_path,
        text=True,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        check=False,
    )
    if check and result.returncode != 0:
        command = " ".join(["dolt", *args])
        raise DoltError(f"{command} failed\nstdout:\n{result.stdout}\nstderr:\n{result.stderr}")
    return result


def ensure_database(db_path: Path, schema_path: Path, *, remote: str | None, init_schema: bool) -> None:
    db_path.mkdir(parents=True, exist_ok=True)
    if not (db_path / ".dolt").exists():
        run_dolt(db_path, ["init"])

    if remote:
        remotes = run_dolt(db_path, ["remote", "-v"], check=False).stdout
        if "origin" not in remotes:
            run_dolt(db_path, ["remote", "add", "origin", remote])

    if init_schema:
        apply_schema(db_path, schema_path)


def apply_schema(db_path: Path, schema_path: Path) -> None:
    sql_text = schema_path.read_text(encoding="utf-8")
    for statement in split_sql_statements(sql_text):
        run_dolt(db_path, ["sql", "-q", statement])


def split_sql_statements(sql_text: str) -> list[str]:
    statements: list[str] = []
    current: list[str] = []
    for raw_line in sql_text.splitlines():
        line = raw_line.strip()
        if not line or line.startswith("--"):
            continue
        current.append(raw_line)
        if line.endswith(";"):
            statements.append("\n".join(current).rstrip(";"))
            current = []
    if current:
        statements.append("\n".join(current))
    return statements


def sql_literal(value: Any) -> str:
    if value is None:
        return "NULL"
    if isinstance(value, bool):
        return "TRUE" if value else "FALSE"
    if isinstance(value, (int, float)):
        return str(value)
    if isinstance(value, (dict, list)):
        value = json.dumps(value, sort_keys=True, separators=(",", ":"))
    escaped = str(value).replace("\\", "\\\\").replace("'", "''")
    return f"'{escaped}'"


def upsert_rows(
    db_path: Path,
    table: str,
    columns: Sequence[str],
    rows: Iterable[dict[str, Any]],
    *,
    update_columns: Sequence[str] | None = None,
    batch_size: int = 200,
) -> int:
    count = 0
    batch: list[dict[str, Any]] = []
    for row in rows:
        batch.append(row)
        if len(batch) >= batch_size:
            count += upsert_batch(db_path, table, columns, batch, update_columns=update_columns)
            batch = []
    if batch:
        count += upsert_batch(db_path, table, columns, batch, update_columns=update_columns)
    return count


def upsert_batch(
    db_path: Path,
    table: str,
    columns: Sequence[str],
    rows: Sequence[dict[str, Any]],
    *,
    update_columns: Sequence[str] | None,
) -> int:
    values = []
    for row in rows:
        values.append("(" + ", ".join(sql_literal(row.get(column)) for column in columns) + ")")

    sql = f"INSERT INTO {table} ({', '.join(columns)}) VALUES\n" + ",\n".join(values)
    updates = list(update_columns or columns)
    if updates:
        sql += "\nON DUPLICATE KEY UPDATE " + ", ".join(f"{column}=VALUES({column})" for column in updates)
    run_dolt(db_path, ["sql", "-q", sql])
    return len(rows)


def commit_and_maybe_push(db_path: Path, *, message: str, push: bool, branch: str | None) -> None:
    status = run_dolt(db_path, ["status"]).stdout
    if is_clean_status(status):
        return

    run_dolt(db_path, ["add", "."])
    committed = run_dolt(db_path, ["commit", "-m", message], check=False)
    if committed.returncode != 0 and "nothing to commit" not in committed.stderr.lower():
        raise DoltError(f"dolt commit failed\nstdout:\n{committed.stdout}\nstderr:\n{committed.stderr}")

    if push:
        target_branch = branch or current_branch(db_path)
        run_dolt(db_path, ["push", "origin", target_branch])


def current_branch(db_path: Path) -> str:
    current = run_dolt(db_path, ["branch"], check=False).stdout
    for line in current.splitlines():
        line = line.strip()
        if line.startswith("* "):
            return line[2:].strip()
    return "main"


def is_clean_status(status: str) -> bool:
    normalized = status.lower()
    clean_markers = [
        "nothing to commit, working tree clean",
        "nothing to commit, working set clean",
        "working tree clean",
        "working set clean",
    ]
    return any(marker in normalized for marker in clean_markers)
