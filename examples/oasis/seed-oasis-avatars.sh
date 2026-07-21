#!/bin/bash

# Seed script for OASIS avatar instances
# Loads the famous characters from Ready Player One via the KubeGame API
#
# Usage: ./examples/oasis/seed-oasis-avatars.sh [API_URL] [NAMESPACE]
# Default API_URL: http://localhost:8082
# Default NAMESPACE: default

API_URL="${1:-http://localhost:8082}"
NAMESPACE="${2:-default}"
GAME="oasis"
BASE="${API_URL}/api/v1/namespaces/${NAMESPACE}/games/${GAME}"
AVATARS="${BASE}/avatars"

create_avatar() {
  local name="$1"
  local data="$2"

  echo "Creating avatar instance: ${name}..."
  response=$(curl -s -w "\n%{http_code}" -X POST "${AVATARS}" \
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

grant_item() {
  local avatar="$1"
  local item="$2"
  local qty="${3:-1}"

  response=$(curl -s -w "\n%{http_code}" -X POST "${AVATARS}/${avatar}/inventory" \
    -H "Content-Type: application/json" \
    -d "{\"itemName\": \"${item}\", \"quantity\": ${qty}}")

  http_code=$(echo "${response}" | tail -1)

  if [ "${http_code}" -eq 200 ]; then
    echo "  Granted: ${item} x${qty}"
  else
    echo "  FAILED grant ${item} (${http_code})"
  fi
}

equip_item() {
  local avatar="$1"
  local item="$2"

  response=$(curl -s -w "\n%{http_code}" -X POST "${AVATARS}/${avatar}/equip" \
    -H "Content-Type: application/json" \
    -d "{\"itemName\": \"${item}\"}")

  http_code=$(echo "${response}" | tail -1)

  if [ "${http_code}" -eq 200 ]; then
    echo "  Equipped: ${item}"
  else
    echo "  FAILED equip ${item} (${http_code})"
  fi
}

echo "============================================"
echo " OASIS Avatar Seed - Ready Player One"
echo " API: ${AVATARS}"
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
  "achievements": ["Copper Key", "Jade Key", "Crystal Key", "Easter Egg"],
  "customizations": {"Race": "Human", "Gender": "Male", "Class": "Gunter"}
}'
grant_item "parzival" "zemeckis-cube"
grant_item "parzival" "delorean"
grant_item "parzival" "leopardon"
grant_item "parzival" "sword-of-the-ba-heer"
grant_item "parzival" "quarter"
equip_item "parzival" "sword-of-the-ba-heer"
echo ""

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
  "achievements": ["Copper Key", "Jade Key", "Crystal Key", "Easter Egg"],
  "customizations": {"Race": "Human", "Gender": "Female", "Class": "Gunter"}
}'
grant_item "art3mis" "akira-motorcycle"
grant_item "art3mis" "black-tiger-sword"
grant_item "art3mis" "scale-mail"
equip_item "art3mis" "black-tiger-sword"
equip_item "art3mis" "scale-mail"
echo ""

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
  "achievements": ["Copper Key", "Jade Key", "Crystal Key", "Easter Egg"],
  "customizations": {"Race": "Orc", "Gender": "Male", "Class": "Warrior"}
}'
grant_item "aech" "iron-giant"
grant_item "aech" "battleaxe"
grant_item "aech" "heavy-plate-armor"
equip_item "aech" "battleaxe"
equip_item "aech" "heavy-plate-armor"
echo ""

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
  "achievements": ["Copper Key", "Jade Key", "Crystal Key"],
  "customizations": {"Race": "Human", "Gender": "Male", "Class": "Warrior"}
}'
grant_item "daito" "ultraman"
grant_item "daito" "masamune-katana"
grant_item "daito" "samurai-yoroi"
equip_item "daito" "masamune-katana"
equip_item "daito" "samurai-yoroi"
echo ""

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
  "achievements": ["Copper Key", "Jade Key", "Crystal Key", "Easter Egg"],
  "customizations": {"Race": "Human", "Gender": "Male", "Class": "Rogue"}
}'
grant_item "shoto" "raideen"
grant_item "shoto" "wakizashi"
grant_item "shoto" "ninja-garb"
equip_item "shoto" "wakizashi"
equip_item "shoto" "ninja-garb"
echo ""

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
  "achievements": ["Copper Key", "Jade Key", "Crystal Key", "Easter Egg"],
  "customizations": {"Race": "Human", "Gender": "Male", "Class": "Mage"}
}'
grant_item "anorak" "robes-of-anorak"
grant_item "anorak" "catalyst"
grant_item "anorak" "unlimited-coins"
equip_item "anorak" "robes-of-anorak"
echo ""

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
  "achievements": [],
  "customizations": {"Race": "Android", "Gender": "Male", "Class": "Rogue"}
}'
grant_item "i-r0k" "orb-of-osuvox"
grant_item "i-r0k" "skull-armor"
grant_item "i-r0k" "plasma-rifle"
equip_item "i-r0k" "skull-armor"
equip_item "i-r0k" "plasma-rifle"
echo ""

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
  "achievements": [],
  "customizations": {"Race": "Human", "Gender": "Male", "Class": "Warrior"}
}'
grant_item "ioi-655321" "mechagodzilla"
grant_item "ioi-655321" "ioi-corporate-armor"
grant_item "ioi-655321" "ioi-railgun"
grant_item "ioi-655321" "corporate-funds"
equip_item "ioi-655321" "ioi-corporate-armor"
equip_item "ioi-655321" "ioi-railgun"
echo ""

echo "============================================"
echo " Seed complete!"
echo "============================================"
echo ""
echo "Verify with: curl -s ${AVATARS} | python3 -m json.tool"
