#!/bin/bash

# Seed script for OASIS avatar instances
# Loads the famous characters from Ready Player One via the KubeGame API
#
# Usage: ./hack/scripts/seed-oasis-avatars.sh [API_URL] [NAMESPACE]
# Default API_URL: http://localhost:8082
# Default NAMESPACE: default

API_URL="${1:-http://localhost:8082}"
NAMESPACE="${2:-default}"
GAME="oasis"
ENDPOINT="${API_URL}/api/v1/namespaces/${NAMESPACE}/games/${GAME}/avatars"

create_avatar() {
  local name="$1"
  local data="$2"

  echo "Creating avatar instance: ${name}..."
  response=$(curl -s -w "\n%{http_code}" -X POST "${ENDPOINT}" \
    -H "Content-Type: application/json" \
    -d "${data}")

  http_code=$(echo "${response}" | tail -1)
  body=$(echo "${response}" | head -1)

  if [ "${http_code}" -eq 201 ]; then
    echo "  OK"
  else
    echo "  FAILED (${http_code}): ${body}"
  fi
}

echo "============================================"
echo " OASIS Avatar Seed - Ready Player One"
echo " API: ${ENDPOINT}"
echo "============================================"
echo ""

# Parzival (Wade Watts)
# Level 10 gunter, found all keys, owns the OASIS
create_avatar "parzival" '{
  "name": "parzival",
  "avatarType": "oasis-avatar",
  "attributes": {
    "strength": "50",
    "intelligence": "90",
    "agility": "65",
    "hitPoints": "210",
    "armorClass": "45",
    "experiencePoints": "1250000",
    "level": "10",
    "charisma": "60"
  },
  "inventory": [
    {"name": "Zemeckis Cube", "type": "Special", "quantity": 1},
    {"name": "DeLorean", "type": "Transport", "quantity": 1},
    {"name": "Leopardon", "type": "Transport", "quantity": 1},
    {"name": "Sword of the Ba Heer", "type": "Equipment", "quantity": 1},
    {"name": "Quarter", "type": "Currency", "quantity": 1}
  ],
  "achievements": ["Copper Key", "Jade Key", "Crystal Key", "Easter Egg"],
  "customizations": {"Race": "Human", "Gender": "Male", "Class": "Gunter"}
}'

# Art3mis (Samantha Cook)
# Famous gunter and blogger, found all keys
create_avatar "art3mis" '{
  "name": "art3mis",
  "avatarType": "oasis-avatar",
  "attributes": {
    "strength": "55",
    "intelligence": "95",
    "agility": "70",
    "hitPoints": "195",
    "armorClass": "40",
    "experiencePoints": "1100000",
    "level": "10",
    "charisma": "75"
  },
  "inventory": [
    {"name": "Akira Motorcycle", "type": "Transport", "quantity": 1},
    {"name": "Black Tiger Sword", "type": "Equipment", "quantity": 1},
    {"name": "Scale Mail", "type": "Equipment", "quantity": 1}
  ],
  "achievements": ["Copper Key", "Jade Key", "Crystal Key", "Easter Egg"],
  "customizations": {"Race": "Human", "Gender": "Female", "Class": "Gunter"}
}'

# Aech (Helen Harris)
# Skilled warrior and vehicle builder
create_avatar "aech" '{
  "name": "aech",
  "avatarType": "oasis-avatar",
  "attributes": {
    "strength": "80",
    "intelligence": "70",
    "agility": "60",
    "hitPoints": "250",
    "armorClass": "55",
    "experiencePoints": "980000",
    "level": "9",
    "charisma": "65"
  },
  "inventory": [
    {"name": "Iron Giant", "type": "Transport", "quantity": 1},
    {"name": "Battleaxe", "type": "Equipment", "quantity": 1},
    {"name": "Heavy Plate Armor", "type": "Equipment", "quantity": 1}
  ],
  "achievements": ["Copper Key", "Jade Key", "Crystal Key", "Easter Egg"],
  "customizations": {"Race": "Orc", "Gender": "Male", "Class": "Warrior"}
}'

# Daito (Toshiro Yoshiaki)
# Japanese gunter, samurai warrior
create_avatar "daito" '{
  "name": "daito",
  "avatarType": "oasis-avatar",
  "attributes": {
    "strength": "75",
    "intelligence": "72",
    "agility": "85",
    "hitPoints": "180",
    "armorClass": "50",
    "experiencePoints": "870000",
    "level": "8",
    "charisma": "45"
  },
  "inventory": [
    {"name": "Ultraman", "type": "Transport", "quantity": 1},
    {"name": "Masamune Katana", "type": "Equipment", "quantity": 1},
    {"name": "Samurai Yoroi", "type": "Equipment", "quantity": 1}
  ],
  "achievements": ["Copper Key", "Jade Key", "Crystal Key"],
  "customizations": {"Race": "Human", "Gender": "Male", "Class": "Warrior"}
}'

# Shoto (Akihide Karatsu)
# Daito's partner, skilled young samurai
create_avatar "shoto" '{
  "name": "shoto",
  "avatarType": "oasis-avatar",
  "attributes": {
    "strength": "60",
    "intelligence": "78",
    "agility": "88",
    "hitPoints": "160",
    "armorClass": "42",
    "experiencePoints": "820000",
    "level": "8",
    "charisma": "55"
  },
  "inventory": [
    {"name": "Raideen", "type": "Transport", "quantity": 1},
    {"name": "Wakizashi", "type": "Equipment", "quantity": 1},
    {"name": "Ninja Garb", "type": "Equipment", "quantity": 1}
  ],
  "achievements": ["Copper Key", "Jade Key", "Crystal Key", "Easter Egg"],
  "customizations": {"Race": "Human", "Gender": "Male", "Class": "Rogue"}
}'

# Anorak (James Halliday)
# Creator of the OASIS, all-powerful wizard avatar
create_avatar "anorak" '{
  "name": "anorak",
  "avatarType": "oasis-avatar",
  "attributes": {
    "strength": "99",
    "intelligence": "99",
    "agility": "99",
    "hitPoints": "999",
    "armorClass": "99",
    "experiencePoints": "9999999",
    "level": "99",
    "charisma": "30"
  },
  "inventory": [
    {"name": "Robes of Anorak", "type": "Equipment", "quantity": 1},
    {"name": "Catalyst", "type": "Special", "quantity": 1},
    {"name": "Unlimited Coins", "type": "Currency", "quantity": 999999}
  ],
  "achievements": ["Copper Key", "Jade Key", "Crystal Key", "Easter Egg"],
  "customizations": {"Race": "Human", "Gender": "Male", "Class": "Mage"}
}'

# i-r0k
# Mercenary and bounty hunter, works for IOI
create_avatar "i-r0k" '{
  "name": "i-r0k",
  "avatarType": "oasis-avatar",
  "attributes": {
    "strength": "70",
    "intelligence": "55",
    "agility": "50",
    "hitPoints": "200",
    "armorClass": "48",
    "experiencePoints": "600000",
    "level": "7",
    "charisma": "25"
  },
  "inventory": [
    {"name": "Orb of Osuvox", "type": "Special", "quantity": 1},
    {"name": "Skull Armor", "type": "Equipment", "quantity": 1},
    {"name": "Plasma Rifle", "type": "Equipment", "quantity": 1}
  ],
  "achievements": [],
  "customizations": {"Race": "Android", "Gender": "Male", "Class": "Rogue"}
}'

# Sorrento / IOI-655321
# Head of IOI operations, the main antagonist
create_avatar "ioi-655321" '{
  "name": "ioi-655321",
  "avatarType": "oasis-avatar",
  "attributes": {
    "strength": "65",
    "intelligence": "60",
    "agility": "45",
    "hitPoints": "190",
    "armorClass": "60",
    "experiencePoints": "500000",
    "level": "7",
    "charisma": "20"
  },
  "inventory": [
    {"name": "MechaGodzilla", "type": "Transport", "quantity": 1},
    {"name": "IOI Corporate Armor", "type": "Equipment", "quantity": 1},
    {"name": "IOI Railgun", "type": "Equipment", "quantity": 1},
    {"name": "Corporate Funds", "type": "Currency", "quantity": 500000}
  ],
  "achievements": [],
  "customizations": {"Race": "Human", "Gender": "Male", "Class": "Warrior"}
}'

echo ""
echo "============================================"
echo " Seed complete!"
echo "============================================"
echo ""
echo "Verify with: curl -s ${ENDPOINT} | python3 -m json.tool"
