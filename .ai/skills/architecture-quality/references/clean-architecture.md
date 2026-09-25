# Clean Architecture Principles

## 1. Dependency Rule
Source code dependencies must point only inward, toward higher-level policies. Nothing in an inner circle can know anything at all about something in an outer circle.

## 2. Layers

### Domain / Entities
- Contains Enterprise-wide business rules.
- Encapsulates the most general and high-level rules.
- Least likely to change when something external changes.

### Use Cases / Application
- Contains application-specific business rules.
- Implements and orchestrates the flow of data to and from the entities.
- Directs entities to use their critical business rules to achieve the goals of the use case.

### Interface Adapters
- A set of adapters that convert data from the format most convenient for the use cases and entities, to the format most convenient for some external agency such as the Database or the Web.
- Controllers, Presenters, Gateways.

### Frameworks and Drivers
- The outermost layer is generally composed of frameworks and tools such as the Database, the Web Framework, etc.
- Keep these things on the outside where they can do little harm.

## 3. Benefits
- **Independent of Frameworks**: The architecture does not depend on the existence of some library of feature-laden software.
- **Testable**: The business rules can be tested without the UI, Database, Web Server, or any other external element.
- **Independent of UI**: The UI can change easily, without changing the rest of the system.
- **Independent of Database**: You can swap out Oracle or SQL Server, for Mongo, BigTable, CouchDB, or something else.
