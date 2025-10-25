#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import csv
import json
import argparse
import sys
from typing import Dict, Any, Tuple
import requests

ENDPOINT_DEFAULT = "http://localhost:18020/api/find-service"

COLUMN_ALIASES = {
    "intent": {"intent", "intencao", "intenção"},
    "service_id": {"service_id", "id"},
    "service_name": {"service_name", "service", "nome_servico", "nome_serviço"},
}

def detect_columns(header: Dict[str, int]) -> Dict[str, str]:
    """
    Mapeia nomes de colunas flexíveis para os canônicos: intent, service_id, service_name.
    """
    normalized = {h.strip().lower(): h for h in header}
    mapping: Dict[str, str] = {}
    for canon, aliases in COLUMN_ALIASES.items():
        for alias in aliases:
            if alias in normalized:
                mapping[canon] = normalized[alias]
                break
        if canon not in mapping:
            raise ValueError(
                f"Coluna obrigatória ausente para '{canon}'. "
                f"Tente usar uma das alternativas: {sorted(list(aliases))}"
            )
    return mapping

def coerce_id(val: Any) -> str:
    """
    Converte o ID para string comparável (evita diferença entre '101' e 101).
    """
    if val is None:
        return ""
    s = str(val).strip()
    # remove .0 comuns de planilhas
    if s.endswith(".0"):
        s = s[:-2]
    return s

def request_service(endpoint: str, intent: str, timeout: float = 5.0) -> Tuple[bool, Dict[str, Any], str]:
    """
    Faz a chamada POST e retorna (ok_http, payload_dict, erro_texto).
    """
    try:
        resp = requests.post(
            endpoint,
            headers={"Content-Type": "application/json"},
            data=json.dumps({"intent": intent}),
            timeout=timeout,
        )
    except requests.RequestException as e:
        return False, {}, f"Erro de rede: {e}"

    try:
        payload = resp.json()
    except ValueError:
        # não é JSON
        return False, {}, f"Resposta não é JSON. Status {resp.status_code}. Corpo: {resp.text[:300]}"

    ok_http = 200 <= resp.status_code < 300
    if not ok_http:
        return False, payload, f"Status HTTP {resp.status_code} com corpo: {json.dumps(payload)[:300]}"
    return True, payload, ""

def validate_response(payload: Dict[str, Any], exp_id: str, exp_name: str) -> Tuple[bool, str]:
    """
    Valida o JSON no formato esperado:
    {
      "success": true,
      "data": {"service_id": ID, "service_name": service}
    }
    """
    if not isinstance(payload, dict):
        return False, "Payload não é objeto JSON."

    if payload.get("success") is not True:
        return False, f"'success' não é true (valor: {payload.get('success')!r})."

    data = payload.get("data")
    if not isinstance(data, dict):
        return False, "'data' ausente ou não é objeto."

    got_id = coerce_id(data.get("service_id"))
    got_name = (data.get("service_name") or "").strip()

    # Comparação case-insensitive para nome; ID deve bater exatamente (após coerção)
    id_ok = got_id == coerce_id(exp_id)
    name_ok = got_name.lower() == (exp_name or "").strip().lower()

    if id_ok and name_ok:
        return True, ""
    problems = []
    if not id_ok:
        problems.append(f"service_id esperado '{exp_id}', obtido '{got_id}'")
    if not name_ok:
        problems.append(f"service_name esperado '{exp_name}', obtido '{got_name}'")
    return False, "; ".join(problems)

def main():
    parser = argparse.ArgumentParser(description="Teste do endpoint /api/find-service com base em um CSV.")
    parser.add_argument("csv_path", help="Caminho do arquivo CSV com: intent, service_id, service_name")
    parser.add_argument("--endpoint", default=ENDPOINT_DEFAULT, help=f"URL do endpoint (padrão: {ENDPOINT_DEFAULT})")
    parser.add_argument("--timeout", type=float, default=5.0, help="Timeout em segundos para cada requisição (padrão: 5.0)")
    parser.add_argument("--delimiter", default=",", help="Delimitador do CSV (padrão: vírgula ',').")
    args = parser.parse_args()

    total = 0
    passed = 0
    failed = 0

    try:
        with open(args.csv_path, "r", encoding="utf-8-sig", newline="") as f:
            reader = csv.DictReader(f, delimiter=args.delimiter)
            if reader.fieldnames is None:
                print("Erro: CSV sem cabeçalho.", file=sys.stderr)
                sys.exit(2)

            try:
                colmap = detect_columns(reader.fieldnames)
            except ValueError as e:
                print(f"Erro nas colunas do CSV: {e}", file=sys.stderr)
                sys.exit(2)

            print(f"Usando endpoint: {args.endpoint}")
            print(f"Casos de teste: {args.csv_path}")
            print("-" * 80)

            for i, row in enumerate(reader, start=1):
                total += 1
                intent = (row.get(colmap["intent"]) or "").strip()
                exp_id = coerce_id(row.get(colmap["service_id"]))
                exp_name = (row.get(colmap["service_name"]) or "").strip()

                if not intent:
                    failed += 1
                    print(f"[{i:04}] ❌ Linha sem 'intent' (pulado).")
                    continue

                ok_http, payload, http_err = request_service(args.endpoint, intent, timeout=args.timeout)
                if not ok_http:
                    failed += 1
                    print(f"[{i:04}] ❌ intent='{intent}': erro HTTP/transport: {http_err}")
                    continue

                is_valid, reason = validate_response(payload, exp_id, exp_name)
                if is_valid:
                    passed += 1
                    print(f"[{i:04}] ✅ intent='{intent}' → OK (service_id={exp_id}, service_name='{exp_name}')")
                else:
                    failed += 1
                    print(f"[{i:04}] ❌ intent='{intent}' → {reason} | payload={json.dumps(payload, ensure_ascii=False)}")

    except FileNotFoundError:
        print(f"Arquivo CSV não encontrado: {args.csv_path}", file=sys.stderr)
        sys.exit(2)
    except Exception as e:
        print(f"Erro ao processar: {e}", file=sys.stderr)
        sys.exit(2)

    print("-" * 80)
    print(f"Total: {total} | Sucessos: {passed} | Falhas: {failed}")
    sys.exit(0 if failed == 0 else 1)

if __name__ == "__main__":
    main()
