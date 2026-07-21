# KubeGame

A Kubernetes operator for building gamification platforms. KubeGame lets you define game mechanics as Custom Resources and manages the runtime through a REST API.

## Architecture

KubeGame uses a two-layer design:

- **CRD Layer** (Schema/Config): Game designers define blueprints via `kubectl apply`. Controllers provision infrastructure and persist schema definitions to PostgreSQL.
- **REST API Layer** (Runtime Data): Game clients create and manage player data through HTTP endpoints with validation against the blueprints.

```
kubectl apply -f game.yaml        POST /api/v1/games/oasis/avatar-instances
         |                                    |
    CRD Controller                       REST API Handler
         |                                    |
    Creates Postgres              Validates against blueprint
    Persists schema                   Persists instance
         |                                    |
         +----------> PostgreSQL <------------+
```

## Custom Resources

| CRD | Purpose |
|-----|---------|
| **Game** | Top-level resource. Provisions a PostgreSQL deployment per game. |
| **World** | Defines worlds/planets within a game. |
| **Avatar** | Blueprint for avatar types. Defines attribute types, inventory types, and achievement types. |

### Avatar as a Blueprint

The Avatar CRD is a generic scaffold, not a concrete avatar. It defines **what types of attributes, inventory, and achievements** an avatar can have. Actual avatar instances are created via the REST API and stored in the database.

```yaml
apiVersion: kubegame.systemcraftsman.com/v1alpha1
kind: Avatar
metadata:
  name: oasis-avatar
spec:
  game: oasis
  type: "Adventurer"
  attributeTypes:
    - name: "strength"
      valueType: "int"
    - name: "intelligence"
      valueType: "int"
  inventoryTypes:
    - name: "Weapon"
      category: "Equipment"
    - name: "Vehicle"
      category: "Transport"
  achievementTypes:
    - name: "Copper Key"
      description: "Found the first key."
```

## REST API

The API server runs on port `8082` alongside the operator.

| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/games/{game}/avatar-instances` | Create an avatar instance |
| `GET` | `/api/v1/games/{game}/avatar-instances` | List all avatar instances |
| `GET` | `/api/v1/games/{game}/avatar-instances/{name}` | Get a specific instance |
| `DELETE` | `/api/v1/games/{game}/avatar-instances/{name}` | Delete an instance |

Instance creation validates against the avatar blueprint: attributes, inventory categories, and achievements must be defined in the corresponding Avatar CRD.

## Quick Start

### Prerequisites

- Go 1.21+
- A Kubernetes cluster (Kind, Minikube, etc.)
- kubectl
- operator-sdk v1.33.0

### Deploy

```bash
# Install CRDs
make install

# Run the operator locally
make run
```

### Create a Game

```bash
kubectl apply -f hack/resources/oasis.yaml
kubectl apply -f hack/resources/oasis-avatar.yaml
kubectl apply -f hack/resources/incipio.yaml
kubectl apply -f hack/resources/archaide.yaml
kubectl apply -f hack/resources/chthonia.yaml
kubectl apply -f hack/resources/middle-earth.yaml
```

### Seed Avatar Instances

```bash
./hack/scripts/seed-oasis-avatars.sh
```

This loads 8 Ready Player One characters (Parzival, Art3mis, Aech, Daito, Shoto, Anorak, i-r0k, IOI-655321) via the API.

### Verify

```bash
curl -s http://localhost:8082/api/v1/games/oasis/avatar-instances | python3 -m json.tool
```

## Gamification Mechanics Roadmap

KubeGame aims to implement all [35 Gamification Mechanics](https://www.epicwinblog.net/2013/10/the-35-gamification-mechanics-toolkit.html) by @victormanriquey:

- [x] World
- [x] Avatar (blueprint + instance API)
- [ ] Area, Customization
- [ ] Equipment, Vanity/Elite Items, Power-ups (ItemCatalog CRD)
- [ ] Currency, Trading
- [ ] Skills/Traits, XP Points (SkillTree CRD)
- [ ] Quest, Tutorial, Special Challenge
- [ ] Levels, Time Events
- [ ] Achievements (grant/revoke API), Leaderboards, Rankings
- [ ] Rewards (fixed/variable/random), Loot Tables, Easter Eggs
- [ ] Guilds, Parties/Teams, Social Graph, Chat
- [ ] PvP, Punishments, Lifejackets, Ambassadors
- [ ] Progress HUDs (dashboard API)

## Contributing

If you are familiar with gamification, Go, and Kubernetes operators, please check out the [issues](https://github.com/SystemCraftsman/KubeGame/issues) to contribute.
