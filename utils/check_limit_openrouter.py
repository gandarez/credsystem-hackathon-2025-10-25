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
print(json.dumps(response.json(), indent=2))
