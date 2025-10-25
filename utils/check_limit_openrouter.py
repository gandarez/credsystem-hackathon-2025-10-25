import os
import requests
import json

authorization = os.getenv("OPENROUTER_API_KEY")

response    =  requests.get(
    url     = "https://openrouter.ai/api/v1/key",
    headers = {
                  "Authorization": f"Bearer {authorization}"
              }
)

data = response.json()
if 'error' in data and 'code' in data['error'] and data['error']['code'] != 200:
    print(f"Error : {data['error']['message']}")
    exit(1)

usage = data['data']['usage']
print(f'$ {usage:.2f} used')

exit(0)
