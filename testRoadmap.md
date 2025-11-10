---
9. Implementation Roadmap

Phase 1: Foundation (Week 1) - PRIORITY
- Add testing dependencies to go.mod
- Create test/testutil/database.go with DB helpers
- Create test/testutil/auth.go with JWT test utilities
- Write first unit test: auth/jwt_test.go
- Write first integration test: food/repository_test.go
- Set up Makefile with test commands

Phase 2: Critical Business Logic (Week 2)
- Test zero2hero_service.go - ALL protocols, ALL phases
- Test recipe/repository.go nutrition calculations
- Test diary/handler.go entry calculations
- Test authentication flows (register, login, middleware)

Phase 3: Repository Layer (Week 3)
- Test all repository CRUD operations
- Test complex queries (aggregations, date ranges)
- Test constraint violations and error handling

Phase 4: HTTP Layer (Week 4)
- Test all 30+ endpoints with mocked repositories
- Test request validation and error responses
- Test JWT middleware integration
- Add integration tests for critical endpoints

Phase 5: CI/CD & Polish (Week 5)
- Set up GitHub Actions workflow
- Add code coverage reporting (Codecov)
- Add pre-commit hooks for running tests
- Document testing practices in README
- Achieve 75%+ overall coverage

---