# Project Requirements: Small ERP System Using Golang for a Professional Portfolio

I want to build a **Small ERP system using Golang as a professional portfolio project** to apply for remote or contract Backend Developer positions.

Help me develop this application step by step, from architecture design to a fully functional application that is ready to be published on GitHub.

## 1. Project Requirements

### 1. Architecture and System Design

Design a scalable, modular, and maintainable application architecture.

Define the following:
- Application architecture and folder structure.
- PostgreSQL database schema and table relationships.
- Business workflows and transaction flows.
- REST API conventions and response formats.
- Error handling and logging strategies.
- Environment configuration and dependency management.

Keep the architecture practical for a portfolio project. Avoid unnecessary complexity while following professional backend development practices.

### 2. Authentication and Authorization

Implement a complete authentication and authorization system, including:
- User login and logout.
- User management.
- Secure password hashing.
- Authentication middleware.
- Role-Based Access Control (RBAC).
- Role and permission management.
- Access restrictions based on user roles and permissions.

Ensure that protected endpoints cannot be accessed without proper authentication and authorization.

### 3. Master Data Management

Implement CRUD functionality for essential ERP master data, including:
- Products and product categories.
- Units of measurement.
- Customers.
- Suppliers.
- Other supporting master data required by the business workflows.

Implement proper validation, database constraints, and consistent API responses.

### 4. Procurement Module

Implement a procurement workflow covering:
- Purchase order creation and management.
- Purchase order item management.
- Purchase order status transitions.
- Goods receipt processing.
- Inventory updates following goods receipt.
- Accounts payable recording.
- Supplier information and transaction history.

Ensure that receiving goods updates inventory correctly and that the same receipt cannot be processed multiple times.

### 5. Inventory Management Module

Implement inventory management functionality, including:
- Stock receipts and stock issues.
- Inventory adjustments.
- Stock movement history.
- Current stock balance.
- Stock transaction validation.
- Inventory movement tracking by product.
- Prevention of invalid stock operations.

Maintain inventory consistency throughout all business transactions.

### 6. Sales Module

Implement a sales workflow covering:
- Sales transaction creation and management.
- Customer management.
- Sales order status transitions.
- Inventory deduction when goods are issued.
- Accounts receivable recording.
- Payment status tracking.
- Sales transaction history.

Ensure that inventory deductions and financial records remain consistent with their corresponding sales transactions.

### 7. Finance and Basic Accounting Module

Implement basic financial management functionality, including:
- Cash receipts and payments.
- Accounts payable management.
- Accounts receivable management.
- Supplier payment processing.
- Customer payment processing.
- Basic journal entries.
- Chart of Accounts (COA).
- General ledger.

Where appropriate, integrate financial records with procurement, inventory, and sales transactions.

Use database transactions to maintain financial consistency and prevent duplicate postings.

### 8. Reporting Module

Provide essential ERP reports, including:
- Sales reports.
- Purchase reports.
- Inventory reports.
- Stock movement reports.
- Accounts payable reports.
- Accounts receivable reports.
- Cash flow reports.
- Profit and loss statements.
- Basic balance sheets.

Ensure that report calculations are consistent with the underlying transactional data.

Document the accounting assumptions and calculation rules used by each financial report. The balance sheet must follow the accounting equation, and financial statements should be based on a consistent accounting model rather than arbitrary calculations.

### 9. Docker Orchestration

Create the necessary Docker configuration to run the entire application.

Include:
- `Dockerfile` for the Golang application.
- `compose.yaml` for orchestrating the application and its dependencies.
- PostgreSQL database configuration.
- Environment variable management.
- Database migration execution.
- Database and application health checks.
- Startup dependency management.
- Persistent database storage.
- Appropriate container networking.
- A secure configuration for local development and deployment.

The application should be runnable using a simple command:

```bash
docker compose up --build
```

Document any additional setup steps required.

Ensure that application startup handles database readiness correctly and does not depend on arbitrary startup delays.

### 10. Testing and Smoke Testing

Implement a comprehensive testing strategy that includes:

**Unit tests**
- Business logic.
- Input validation.
- Authentication and authorization rules.
- Financial calculations.
- Inventory calculations.
- Report calculations.

**Integration tests**
- PostgreSQL database operations.
- Database migrations.
- API endpoints.
- Transaction handling.
- Procurement workflows.
- Sales workflows.
- Payment processing.
- Financial record creation.

**Smoke tests**

Create automated smoke tests to verify that the application is operational after deployment.

At a minimum, verify:
- Application health endpoints.
- Database connectivity.
- Login and authentication.
- Protected endpoint access.
- Master data operations.
- Purchase order creation.
- Goods receipt processing.
- Inventory balance updates.
- Sales transactions.
- Payment processing.
- Financial report generation.

Also test important failure scenarios, including invalid input, insufficient stock, unauthorized access, duplicate transaction processing, and database constraint violations.

**Important:** Actually execute the tests, inspect their results, fix any identified issues, and rerun the relevant tests. Do not claim that a test passed unless it has been executed successfully.

### 11. Professional Documentation

Create a comprehensive and professional `README.md` suitable for a public GitHub portfolio.

Include:
- Project overview.
- Business problems addressed.
- Key features.
- Technology stack.
- Architecture diagram.
- Database ERD.
- Project folder structure.
- Installation instructions.
- Environment configuration.
- Docker setup.
- Database migration instructions.
- API documentation.
- Authentication and authorization details.
- Example API requests and responses.
- Demo data and demo account instructions.
- Testing instructions.
- Smoke test execution.
- Deployment instructions.
- Known limitations and future improvements.

Ensure that another developer can clone the repository, configure the environment, start the application, and run the tests by following the documentation.

### 12. Final Validation and Portfolio Readiness

Before considering the project complete, verify that:

1. The application can be built successfully.
2. All required Docker containers start correctly.
3. Database migrations execute successfully.
4. Initial seed data can be loaded.
5. Authentication and authorization work as expected.
6. All major CRUD operations function correctly.
7. Procurement transactions update inventory and financial records correctly.
8. Sales transactions update inventory and financial records correctly.
9. Payment processing updates outstanding balances correctly.
10. Reports produce consistent results based on the underlying data.
11. Automated tests pass.
12. Smoke tests pass against the running application.
13. The application can be started from a clean environment using the documented instructions.
14. No credentials, secrets, or sensitive information are exposed in the repository.
15. The GitHub repository is organized, documented, and ready for technical evaluation.

If any issues are discovered, fix them and rerun the relevant tests before declaring the project ready.

---

## 2. Technical Requirements and Development Principles

Follow these development principles throughout the project:

- **Primary technologies:** Golang and PostgreSQL.
- Use a modular architecture with clear separation of concerns.
- Follow Go conventions and established backend development practices.
- Use RESTful API design principles.
- Implement proper input validation and structured error responses.
- Use secure password hashing and appropriate authentication mechanisms.
- Implement RBAC for authorization.
- Use database transactions wherever multiple related database operations must succeed or fail together.
- Enforce data integrity through appropriate database constraints.
- Prevent duplicate transaction processing and unintended repeated financial postings.
- Avoid hardcoded credentials, environment-specific configuration, and business values that should be configurable.
- Use database migrations to manage schema changes.
- Provide seed data for demonstration purposes.
- Pin or document compatible dependency versions.
- Include structured logging and appropriate health checks.
- Keep the code readable, maintainable, and easy to review.
- Avoid unnecessary dependencies and overengineering.
- Never fabricate test results or claim successful validation without evidence.

Prioritize correctness, security, maintainability, and practical business functionality over unnecessary complexity.

## 3. Development Workflow

Develop the project incrementally rather than generating the entire application in one step.

Follow this workflow:

**Phase 1 — Planning and Architecture**
- Define the MVP scope.
- Identify the core business entities.
- Design the database schema and ERD.
- Define module boundaries and business workflows.
- Establish API conventions and the project structure.

**Phase 2 — Foundation**
- Initialize the Golang project.
- Configure PostgreSQL connectivity.
- Implement database migrations.
- Configure logging, environment variables, and error handling.
- Set up Docker and Docker Compose.

**Phase 3 — Core Business Modules**
- Implement authentication and authorization.
- Implement master data management.
- Implement procurement and inventory workflows.
- Implement sales workflows.
- Implement finance and basic accounting.

**Phase 4 — Reporting**
- Implement operational reports.
- Implement financial reports.
- Validate report calculations against transactional data.

**Phase 5 — Testing and Validation**
- Add unit and integration tests.
- Execute database migration tests.
- Run API tests.
- Execute end-to-end business workflow tests.
- Run smoke tests against the containerized application.
- Fix discovered problems and repeat the relevant tests.

**Phase 6 — Documentation and Deployment**
- Complete the README.
- Add architecture and ERD diagrams.
- Prepare demo data.
- Validate the project from a clean environment.
- Prepare the repository for GitHub publication.
- Document deployment options and any limitations.

At each phase, explain the implementation decisions, identify the files being created or modified, and verify that the current phase works before proceeding.

## 4. Final Objective

The final deliverable must be a **fully functional, containerized Small ERP application built with Golang and PostgreSQL**, featuring authentication, master data management, procurement, inventory, sales, finance, accounting, and reporting.

The project must include Docker orchestration, database migrations, seed data, automated tests, smoke tests, architecture documentation, and clear setup instructions.

It should demonstrate practical backend engineering skills, including API development, database design, transaction management, business logic implementation, containerization, testing, and documentation.

The repository should be suitable for a professional GitHub portfolio and help demonstrate my qualifications when applying for remote Backend Developer positions or contract opportunities.

**Start by proposing a realistic MVP scope and system architecture. Then implement the project step by step, validate each module, and continue until the application can be built, started, and tested successfully using the documented Docker workflow.**

Prioritize delivering a working, verifiable application over simply generating a large amount of code.