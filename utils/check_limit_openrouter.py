import requests
import json
import os
key = os.environ["OPENROUTER_API_KEY"]

response = requests.get(
  url="https://openrouter.ai/api/v1/key",
  headers={
    "Authorization": f"Bearer {key}"
  }
)
data = response.json()
data["data"]["usage"]
print(f'${data["data"]["usage"]:.2f} used today.')
