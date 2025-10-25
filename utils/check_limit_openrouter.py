import requests
import json
response = requests.get(
  url="https://openrouter.ai/api/v1/key",
  headers={
    "Authorization": f"Bearer "
  }
)
data = response.json()
data["data"]["usage"]
print(f'${data["data"]["usage"]:.2f} used today.')
