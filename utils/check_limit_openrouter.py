import os
import requests
import json
from dotenv import load_dotenv


GET_PATH = os.path.join(os.path.dirname(__file__), '..', 'participantes', 'controladores_do_panico_16', '.env')
load_dotenv(dotenv_path=GET_PATH)
TOKEN = os.getenv("OPENROUTER_API_KEY")

response = requests.get(
  url="https://openrouter.ai/api/v1/key",
  headers={
    "Authorization": f"Bearer {TOKEN}"
  }
)
data = response.json()
data["data"]["usage"]
print(f'${data["data"]["usage"]:.2f} used today.')
