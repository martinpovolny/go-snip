# Personal Data — Students & Employees

**Auth**: HTTP Basic

## Students

**Source name**: Osobní údaje žáků  
**Base path**: `/api/student-data/v1/`

Lists personal data for students enrolled at the school. Which fields are returned depends on the API module configuration in Edookit admin.

### List Students

```
GET /api/student-data/v1/list[/{date}]
```

#### Parameters

| Name | Type | Description |
|------|------|-------------|
| `{date}` | string (path) | Reference date `YYYY-MM-DD`. Defaults to today. |
| `include-inactive-since` | string (query) | Include students who left school after this date. |

#### Response Fields

Returns `{"Students": [...]}`.

| Field | Type | Nullable | Description |
|-------|------|----------|-------------|
| `PersonId` | int | no | Edookit person ID |
| `Firstname` | string | no | First name |
| `Middlename` | string | yes | Middle name |
| `Lastname` | string | no | Last name |
| `DegreePreceding` | string | yes | Title before name |
| `DegreeFollowing` | string | yes | Title after name |
| `DateOfBirth` | string | yes | `YYYY-MM-DD` |
| `Age` | int | yes | Age at reference date |
| `Gender` | string | yes | `"M"`, `"F"`, or null |
| `PermanentStayAddress` | object | yes | Address with `CountryCode`, `PostalCode`, `City`, `Street`, `HouseNo`, `LandRegistryNo` |
| `CitizenshipCountryCode` | string | yes | 3-letter ISO country code |
| `HealthInsuranceCompanyCode` | string | yes | Health insurance code |
| `OrganizationName` | string | no | School name |
| `EnrolledSince` | string | no | `YYYY-MM-DD` |
| `UnenrolledSince` | string | yes | `YYYY-MM-DD` or null if still enrolled |
| `ClassName` | string | yes | Class at reference date |
| `LearningGroupNames` | string[] | no | Learning groups |
| `CurrentGradeNum` | int | yes | Grade number at reference date |
| `Phone` | string | yes | Phone |
| `PhoneMobile` | string | yes | Mobile |
| `PrimaryEmail` | string | yes | Email |
| `Plus4UId` | string | yes | Plus4U account ID |
| `VariableSymbol` | string | yes | Variable symbol (for payments) |

---

## Employees

**Source name**: Osobní údaje zaměstnanců  
**Base path**: `/api/employee-data/v1/`

### List Employees

```
GET /api/employee-data/v1/list[/{date}]
```

Same parameters as students. Returns `{"Employees": [...]}` with similar fields plus:

| Field | Type | Description |
|-------|------|-------------|
| `NameAbbr` | string | Name abbreviation |
| `CustomIdentifier` | string | Custom employee ID |
| `FamilyStatus` | string | `"S"` single, `"M"` married, `"D"` divorced, `"W"` widowed |
| `ClassNames` | string[] | Classes this employee is class teacher of |
