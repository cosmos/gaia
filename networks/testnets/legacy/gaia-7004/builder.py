import json
from hashlib import sha256
from base64 import b64decode

j = json.loads(open("genesis.json","r").read())

#filter out bonds to no longer existing validators
j['app_state']['stake']['bonds'] = [b for b in j['app_state']['stake']['bonds'] if b['validator_addr'] in [v['owner'] for v in j['app_state']['stake']['validators']]]



#update main validator section to match validators from state machine
j["validators"] = [{"pub_key":v["pub_key"], "power":v["tokens"], "name":v["description"]["moniker"]} for v in j["app_state"]["stake"]["validators"]]


coinsToMap = lambda x: type(x)==list and {c["denom"]:c["amount"] for c in x} or {}

#update pool total values to be accurate
j["app_state"]["stake"]["pool"]["loose_tokens"] = str(sum([int(coinsToMap(a["coins"]).get("steak",0)) for a in j["app_state"]["accounts"]]))
j["app_state"]["stake"]["pool"]["bonded_tokens"] = str(sum([int(v["power"]) for v in j["validators"]]))

open("7004-genesis.json","w").write(json.dumps(j, indent=2, sort_keys=True))
