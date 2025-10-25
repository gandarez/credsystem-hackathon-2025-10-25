import requests
import json
response = requests.get(
  url="https://openrouter.ai/api/v1/key",
  headers={
    "Authorization": f"Bearer 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"
  }
)
print(json.dumps(response.json(), indent=2))
