#!/usr/bin/env python3
"""
Exemplo de teste manual para demonstrar a validação
"""

import json
import requests

def test_single_intent():
    """Testa um intent individual para demonstrar a validação"""
    
    # Dados de exemplo do CSV
    intent_data = {
        'service_id': '9',
        'service_name': 'Desbloqueio de Cartão',
        'intent': 'desbloquear cartão'
    }
    
    # Fazer requisição
    api_url = "http://localhost:18020/api/intent"
    payload = {"intent": intent_data['intent']}
    
    print(f"🧪 Testando intent: '{intent_data['intent']}'")
    print(f"📋 Esperado - service_id: {intent_data['service_id']}, service_name: '{intent_data['service_name']}'")
    print()
    
    try:
        response = requests.post(api_url, json=payload, headers={'Content-Type': 'application/json'})
        
        print(f"📡 Status HTTP: {response.status_code}")
        
        if response.status_code == 200:
            response_data = response.json()
            print(f"📄 Resposta JSON:")
            print(json.dumps(response_data, indent=2, ensure_ascii=False))
            print()
            
            # Validação
            if response_data.get('success'):
                data = response_data.get('data', {})
                actual_service_id = data.get('service_id')
                actual_service_name = data.get('service_name')
                
                print("🔍 VALIDAÇÃO:")
                print(f"   service_id: esperado {intent_data['service_id']}, recebido {actual_service_id}")
                print(f"   service_name: esperado '{intent_data['service_name']}', recebido '{actual_service_name}'")
                
                service_id_match = str(actual_service_id) == intent_data['service_id']
                service_name_match = actual_service_name == intent_data['service_name']
                
                print(f"   ✅ service_id match: {service_id_match}")
                print(f"   ✅ service_name match: {service_name_match}")
                
                if service_id_match and service_name_match:
                    print("   🎉 VALIDAÇÃO PASSOU!")
                else:
                    print("   ❌ VALIDAÇÃO FALHOU!")
            else:
                print(f"❌ API retornou success: false - {response_data.get('error', 'Erro desconhecido')}")
        else:
            print(f"❌ Erro HTTP: {response.text}")
            
    except Exception as e:
        print(f"❌ Erro: {e}")

if __name__ == "__main__":
    print("🧪 Exemplo de Teste de Validação")
    print("=" * 50)
    test_single_intent()
