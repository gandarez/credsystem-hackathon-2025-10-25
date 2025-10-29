#!/usr/bin/env python3
"""
Script para testar e validar a lista de intents contra a API
"""

import csv
import json
import requests
import time
from typing import List, Dict, Tuple
import sys

# Configurações
API_URL = "http://localhost:18020/api/intent"
CSV_FILE = "assets/intents-pre-loaded.csv"
DELAY_BETWEEN_REQUESTS = 0.5  # Delay em segundos entre requisições

class IntentTester:
    def __init__(self, api_url: str, csv_file: str):
        self.api_url = api_url
        self.csv_file = csv_file
        self.results = []
        self.stats = {
            'total': 0,
            'success': 0,
            'error': 0,
            'timeout': 0,
            'validation_failures': 0
        }
    
    def read_intents_from_csv(self) -> List[Dict[str, str]]:
        """Lê os intents do arquivo CSV"""
        intents = []
        try:
            with open(self.csv_file, 'r', encoding='utf-8') as file:
                reader = csv.DictReader(file, delimiter=';')
                for row in reader:
                    if row['intent'].strip():  # Ignora linhas vazias
                        intents.append({
                            'service_id': row['service_id'],
                            'service_name': row['service_name'],
                            'intent': row['intent'].strip()
                        })
        except FileNotFoundError:
            print(f"❌ Erro: Arquivo {self.csv_file} não encontrado!")
            sys.exit(1)
        except Exception as e:
            print(f"❌ Erro ao ler arquivo CSV: {e}")
            sys.exit(1)
        
        return intents
    
    def validate_service_data(self, intent_data: Dict[str, str], response_data: Dict) -> Dict:
        """Valida se service_id e service_name retornados correspondem aos esperados"""
        validation = {
            'service_id_match': False,
            'service_name_match': False,
            'validation_error': None
        }
        
        # Verificar se existe data na resposta
        if 'data' not in response_data or response_data['data'] is None:
            validation['validation_error'] = 'Resposta não contém dados (data é null)'
            return validation
        
        data = response_data['data']
        
        # Validar service_id
        expected_service_id = int(intent_data['service_id'])
        actual_service_id = data.get('service_id')
        
        if actual_service_id is None:
            validation['validation_error'] = 'service_id não encontrado na resposta'
            return validation
        
        try:
            actual_service_id = int(actual_service_id)
            validation['service_id_match'] = (expected_service_id == actual_service_id)
        except (ValueError, TypeError):
            validation['validation_error'] = f'service_id inválido na resposta: {actual_service_id}'
            return validation
        
        # Validar service_name
        expected_service_name = intent_data['service_name']
        actual_service_name = data.get('service_name')
        
        if actual_service_name is None:
            validation['validation_error'] = 'service_name não encontrado na resposta'
            return validation
        
        validation['service_name_match'] = (expected_service_name == actual_service_name)
        
        # Gerar mensagem de erro se necessário
        if not validation['service_id_match'] or not validation['service_name_match']:
            errors = []
            if not validation['service_id_match']:
                errors.append(f"service_id: esperado {expected_service_id}, recebido {actual_service_id}")
            if not validation['service_name_match']:
                errors.append(f"service_name: esperado '{expected_service_name}', recebido '{actual_service_name}'")
            validation['validation_error'] = '; '.join(errors)
        
        return validation
    
    def test_intent(self, intent_data: Dict[str, str]) -> Dict:
        """Testa um intent individual na API"""
        payload = {
            "intent": intent_data['intent']
        }
        
        try:
            response = requests.post(
                self.api_url,
                json=payload,
                headers={'Content-Type': 'application/json'},
                timeout=30
            )
            
            result = {
                'service_id': intent_data['service_id'],
                'service_name': intent_data['service_name'],
                'intent': intent_data['intent'],
                'status_code': response.status_code,
                'success': False,
                'error': None,
                'response_data': None,
                'validation': {
                    'service_id_match': False,
                    'service_name_match': False,
                    'validation_error': None
                }
            }
            
            if response.status_code == 200:
                try:
                    response_data = response.json()
                    result['response_data'] = response_data
                    result['success'] = response_data.get('success', False)
                    
                    if result['success']:
                        # Validar service_id e service_name
                        validation_result = self.validate_service_data(intent_data, response_data)
                        result['validation'] = validation_result
                        
                        # Se a validação falhou, marcar como erro
                        if not validation_result['service_id_match'] or not validation_result['service_name_match']:
                            result['success'] = False
                            result['error'] = validation_result['validation_error']
                    else:
                        result['error'] = response_data.get('error', 'Erro desconhecido')
                except json.JSONDecodeError:
                    result['error'] = 'Resposta não é um JSON válido'
            else:
                result['error'] = f'HTTP {response.status_code}: {response.text}'
            
            return result
            
        except requests.exceptions.Timeout:
            return {
                'service_id': intent_data['service_id'],
                'service_name': intent_data['service_name'],
                'intent': intent_data['intent'],
                'status_code': None,
                'success': False,
                'error': 'Timeout da requisição',
                'response_data': None
            }
        except requests.exceptions.ConnectionError:
            return {
                'service_id': intent_data['service_id'],
                'service_name': intent_data['service_name'],
                'intent': intent_data['intent'],
                'status_code': None,
                'success': False,
                'error': 'Erro de conexão com a API',
                'response_data': None
            }
        except Exception as e:
            return {
                'service_id': intent_data['service_id'],
                'service_name': intent_data['service_name'],
                'intent': intent_data['intent'],
                'status_code': None,
                'success': False,
                'error': f'Erro inesperado: {str(e)}',
                'response_data': None
            }
    
    def run_tests(self):
        """Executa todos os testes"""
        print("🚀 Iniciando testes de intents...")
        print(f"📡 API URL: {self.api_url}")
        print(f"📄 Arquivo CSV: {self.csv_file}")
        print("-" * 60)
        
        # Ler intents do CSV
        intents = self.read_intents_from_csv()
        self.stats['total'] = len(intents)
        
        print(f"📊 Total de intents encontrados: {self.stats['total']}")
        print()
        
        # Testar cada intent
        for i, intent_data in enumerate(intents, 1):
            print(f"[{i:3d}/{self.stats['total']}] Testando: '{intent_data['intent']}'")
            
            result = self.test_intent(intent_data)
            self.results.append(result)
            
            # Atualizar estatísticas
            if result['success']:
                self.stats['success'] += 1
                print(f"    ✅ Sucesso")
            else:
                if 'timeout' in result['error'].lower():
                    self.stats['timeout'] += 1
                elif 'validation' in result.get('error', '').lower() or 'service_id' in result.get('error', '').lower() or 'service_name' in result.get('error', '').lower():
                    self.stats['validation_failures'] += 1
                else:
                    self.stats['error'] += 1
                print(f"    ❌ Erro: {result['error']}")
            
            # Delay entre requisições
            if i < self.stats['total']:
                time.sleep(DELAY_BETWEEN_REQUESTS)
        
        print()
        self.print_summary()
        self.save_detailed_report()
    
    def print_summary(self):
        """Imprime resumo dos resultados"""
        print("=" * 60)
        print("📈 RESUMO DOS TESTES")
        print("=" * 60)
        print(f"Total de intents testados: {self.stats['total']}")
        print(f"✅ Sucessos: {self.stats['success']} ({self.stats['success']/self.stats['total']*100:.1f}%)")
        print(f"❌ Erros gerais: {self.stats['error']} ({self.stats['error']/self.stats['total']*100:.1f}%)")
        print(f"🔍 Falhas de validação: {self.stats['validation_failures']} ({self.stats['validation_failures']/self.stats['total']*100:.1f}%)")
        print(f"⏱️  Timeouts: {self.stats['timeout']} ({self.stats['timeout']/self.stats['total']*100:.1f}%)")
        print()
        
        # Mostrar erros por categoria de serviço
        errors_by_service = {}
        for result in self.results:
            if not result['success']:
                service_name = result['service_name']
                if service_name not in errors_by_service:
                    errors_by_service[service_name] = 0
                errors_by_service[service_name] += 1
        
        if errors_by_service:
            print("❌ Erros por categoria de serviço:")
            for service, count in sorted(errors_by_service.items()):
                print(f"   {service}: {count} erros")
        
        # Mostrar falhas de validação específicas
        validation_failures = [r for r in self.results if not r['success'] and 'validation' in r.get('error', '').lower()]
        if validation_failures:
            print("\n🔍 Falhas de validação específicas:")
            for failure in validation_failures[:5]:  # Mostrar apenas as primeiras 5
                print(f"   Intent: '{failure['intent']}'")
                print(f"   Erro: {failure['error']}")
                print()
    
    def save_detailed_report(self):
        """Salva relatório detalhado em JSON"""
        report = {
            'summary': self.stats,
            'api_url': self.api_url,
            'csv_file': self.csv_file,
            'timestamp': time.strftime('%Y-%m-%d %H:%M:%S'),
            'results': self.results
        }
        
        filename = f"intent_test_report_{int(time.time())}.json"
        with open(filename, 'w', encoding='utf-8') as f:
            json.dump(report, f, indent=2, ensure_ascii=False)
        
        print(f"📄 Relatório detalhado salvo em: {filename}")

def main():
    print("🧪 Testador de Intents - API de Classificação")
    print("=" * 60)
    
    # Verificar se a API está rodando
    try:
        response = requests.get("http://localhost:18020/healthz", timeout=5)
        if response.status_code == 200:
            print("✅ API está rodando e acessível")
        else:
            print("⚠️  API respondeu mas com status inesperado")
    except:
        print("❌ API não está acessível em http://localhost:18020")
        print("   Certifique-se de que o servidor está rodando!")
        sys.exit(1)
    
    print()
    
    # Executar testes
    tester = IntentTester(API_URL, CSV_FILE)
    tester.run_tests()

if __name__ == "__main__":
    main()
