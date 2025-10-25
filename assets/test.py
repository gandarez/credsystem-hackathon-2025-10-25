#!/usr/bin/env python3
import argparse
import csv
import json
import time
from collections import defaultdict
from pathlib import Path

import requests

def parse_args():
    ap = argparse.ArgumentParser(
        description="Bate na API /api/find-service com intents do CSV e reporta esperado vs obtido."
    )
    ap.add_argument(
        "--csv",
        default="intents_pre_loaded.csv",
        help="Caminho do CSV (padrão: intents_pre_loaded.csv)",
    )
    ap.add_argument(
        "--url",
        default="http://localhost:8080/api/find-service",
        help="URL da API (padrão: http://localhost:8080/api/find-service)",
    )
    ap.add_argument(
        "--timeout",
        type=float,
        default=5.0,
        help="Timeout de requisição em segundos (padrão: 5.0)",
    )
    ap.add_argument(
        "--max",
        type=int,
        default=0,
        help="Máximo de linhas para testar (0 = todas).",
    )
    return ap.parse_args()

def extract_pred(response_text, response_json):
    """
    Tenta extrair o service_id da resposta.
    Suporta:
      - body com apenas um inteiro: "7"
      - JSON com {"service_id":7} ou {"id":7} ou {"result":{"service_id":7}}
    """
    # 1) tentar JSON estruturado
    if response_json is not None:
        # caminhos comuns
        for key in ("service_id", "id"):
            if key in response_json and isinstance(response_json[key], int):
                return response_json[key]
        # tentar um nível abaixo (ex.: {"result": {"service_id": 7}})
        for subkey in ("result", "data", "output"):
            if subkey in response_json and isinstance(response_json[subkey], dict):
                sub = response_json[subkey]
                for key in ("service_id", "id"):
                    if key in sub and isinstance(sub[key], int):
                        return sub[key]
    # 2) tentar inteiro puro no body
    txt = response_text.strip()
    num = 0
    if txt and txt[0].isdigit():
        for ch in txt:
            if ch.isdigit():
                num = num * 10 + (ord(ch) - 48)
            else:
                break
        return num
    # 3) falhou → 0
    return 0

def main():
    args = parse_args()
    csv_path = Path(args.csv)
    if not csv_path.exists():
        raise SystemExit(f"CSV não encontrado: {csv_path}")

    # leitura do CSV com separador ';'
    rows = []
    with csv_path.open("r", encoding="utf-8-sig", newline="") as f:
        reader = csv.DictReader(f, delimiter=";")
        # suportar cabeçalhos: service_id;service_name;intent
        for i, row in enumerate(reader, start=2):  # start=2 (desconta header na 1)
            try:
                sid = int(row["service_id"])
            except Exception:
                # pular linhas inconsistentes
                continue
            rows.append({
                "lineno": i,
                "service_id": sid,
                "service_name": row.get("service_name", "").strip(),
                "intent": row.get("intent", "").strip(),
            })
            if args.max and len(rows) >= args.max:
                break

    if not rows:
        print("Nenhuma linha válida no CSV.")
        return

    session = requests.Session()
    ok = 0
    total = 0
    conf = defaultdict(int)  # (esperado, obtido) -> contagem

    print("linha | esperado | obtido | ms   | intent (resumo)")
    print("-" * 90)

    for r in rows:
        payload = {"intent": r["intent"]}
        t0 = time.perf_counter()
        try:
            resp = session.post(args.url, json=payload, timeout=args.timeout)
            dt = (time.perf_counter() - t0) * 1000.0
        except Exception as e:
            dt = (time.perf_counter() - t0) * 1000.0
            pred = -1  # erro de requisição
            print(f"{r['lineno']:5d} | {r['service_id']:8d} | {pred:6d} | {dt:4.0f} | ERRO: {e}")
            conf[(r["service_id"], pred)] += 1
            total += 1
            continue

        text = resp.text
        try:
            js = resp.json()
        except Exception:
            js = None

        pred = extract_pred(text, js)
        conf[(r["service_id"], pred)] += 1

        total += 1
        if pred == r["service_id"]:
            ok += 1

        # intent resumido (até 40 chars)
        it = r["intent"].replace("\n", " ")
        if len(it) > 40:
            it = it[:37] + "..."

        print(f"{r['lineno']:5d} | {r['service_id']:8d} | {pred:6d} | {dt:4.0f} | {it}")

    print("\nResumo")
    print("------")
    acc = ok / total * 100.0
    print(f"Total: {total}  |  Corretos: {ok}  |  Acurácia: {acc:.2f}%")

    # Confusion-like (apenas erros)
    errors = [(k, v) for k, v in conf.items() if k[0] != k[1]]
    if errors:
        print("\nErros (esperado -> obtido : contagem):")
        # ordena por frequência
        errors.sort(key=lambda kv: kv[1], reverse=True)
        for (exp_, got_), cnt in errors[:30]:
            print(f"{exp_:>3} -> {got_:>3} : {cnt}")

if __name__ == "__main__":
    main()
