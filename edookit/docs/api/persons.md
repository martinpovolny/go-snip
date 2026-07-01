# Vyhledání osob (Person Search)

## Configuration

| Field | Value |
|-------|-------|
| Source name | Vyhledávání osob |
| Domain | `https://<school>.edookit.net` |
| Auth | HTTP Basic |

## Endpoints

### Get person by ID

```
GET /api/person-search/v1/{searchCriterion}
```

Returns one person matching the criterion, plus their children and representatives.

**Path parameter**

`searchCriterion` must be an **Edookit ID** (integer, e.g. `237`) or a **Plus4U ID**
(e.g. `123-4432-1`). A plain name string is not valid and returns `400`.

**Response (200)** — always an array

```json
[
  {
    "firstname": "Jana",
    "lastname": "Slepičková",
    "edookit_id": 165,
    "email": "superhen@seznam.cz",
    "roles": ["parent", "employee"],
    "children": [
      {
        "legalrole": "legal_representative",
        "edookit_id": 52,
        "firstname": "Martin",
        "lastname": "Slepička",
        "email": "kohout@gmail.com",
        "plus4u_id": null,
        "class": "5.A"
      }
    ],
    "representatives": [],
    "plus4u_id": "123-4432-1",
    "class": null
  }
]
```

| Field | Type | Notes |
|-------|------|-------|
| `edookit_id` | int | Unique within this school instance only |
| `roles` | string[] | `student`, `parent`, `employee`; empty for former members |
| `children` | object[] | Top-level result only, not nested |
| `representatives` | object[] | Top-level result only, not nested |
| `legalrole` | string | Only inside `children`/`representatives` |
| `class` | string\|null | Populated for students only |

Empty array = valid format but person not found.

## Go usage

```go
c := edookit.New("https://yourschool.edookit.net", user, pass)

persons, err := c.GetPerson(237)          // by integer Edookit ID
persons, err := c.SearchPerson("123-4432-1") // by Plus4U ID
```

## curl

```bash
bin/api /api/person-search/v1/237
```
