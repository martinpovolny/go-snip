# Attendance

**Source name**: Docházka  
**Auth**: HMAC (special — see below)  
**Base path**: `/api/dochazka/v2/`

> **Vendor registration required**: Unlike other modules, the Attendance module requires the vendor system to be registered with Edookit. Contact support@edookit.com with: company name, attendance system name, business contact, developer contact.

## Authentication

Each request must include three custom HTTP headers:

| Header | Value |
|--------|-------|
| `com.edookit.Client` | Client application identifier (`clientId`) provided by Edookit |
| `com.edookit.Auth` | `<mark>:<hmac>` — mark (school username) + HMAC-SHA1 signature |
| `com.edookit.Time` | Request timestamp |

### Timestamp Format

- `yyyy-mm-dd hh:mm:ss.sss` (local CET/CEST, up to 6 decimal places for seconds)
- ISO-8601

Must be within ±15 minutes of server time. The same timestamp cannot be reused (replay protection).

### HMAC Calculation

```
HMAC-SHA1(clientKey, "METHOD+PATH+TIMESTAMP+PASSWORD")
```

- `clientKey` — provided by Edookit at registration
- `METHOD` — HTTP method (e.g. `GET`)
- `PATH` — Request path (e.g. `/api/dochazka/v2/zaci/1234`)
- `TIMESTAMP` — same as `com.edookit.Time` header
- `PASSWORD` — school-specific password set by admin in Edookit

The auth header value is `<mark>:<hex_hmac>`.

## Endpoints

### Version (Public)
```
GET /api/dochazka/v2/verze
```
Returns `{"VerzeRozhrani": "2.10.0", "Zdroj": "EDOOKIT"}`.

### School Settings
```
GET /api/dochazka/v2/nastaveni
```
Returns current school year, semester, school name and address.

### List Classes
```
GET /api/dochazka/v2/tridy[/{pk}]
```
Returns `{"Tridy": [{"PkTrida", "Zkratka", "Rocnik", "PkTridniUcitel", "EvSkupina"}]}`.

### Schedule Groups / Courses
```
GET /api/dochazka/v2/klas-skupiny[/{pk}]
```
Positive `PkKlasSkupina` = schedule group; negative = course.

### List Workers
```
GET /api/dochazka/v2/pracovnici[/{pk}]
```

### List Students
```
GET /api/dochazka/v2/zaci[/{pk}]
```
Includes legal representative data (Z1*, Z2* fields).

### Daily Schedule for a Person
```
GET /api/dochazka/v2/rozvrh/{datum}/osoba/{pk}
GET /api/dochazka/v2/rozvrh/{datum}/pracovnik/{pk}
GET /api/dochazka/v2/rozvrh/{datum}/zak/{pk}
```
Returns `{"CasovyPlan": [...]}` — timetable entries for the given date.

### Read Passes
```
GET /api/dochazka/v2/pruchody/den/{datum}
GET /api/dochazka/v2/pruchody/{pk}
```
Returns door passage records (entry/exit times).

### Insert Pass
```
POST /api/dochazka/v2/pruchody
```
Body (JSON or XML): `PkUzivatel`, `Datum`, `Cas`, `Smer` (`"P"` in / `"O"` out / `"X"` no direction), `Hlavni`, `BranaId`, `CteckaId`, `Poznamka`.  
Returns HTTP 201 on success with `Location` header.
