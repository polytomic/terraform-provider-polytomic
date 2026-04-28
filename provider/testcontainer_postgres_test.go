package provider

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testPGUser     = "polytomic"
	testPGPassword = "polytomic"
	testPGDatabase = "polytomic"
)

type pgContainer struct {
	host string
	port int
}

var (
	sharedPGContainer     *pgContainer
	sharedPGContainerOnce sync.Once
	sharedPGContainerErr  error
)

// getSharedPGContainer starts a Postgres testcontainer once per test run and
// returns the host/port to reach it. The container is started with the schemas
// and seed data required by sync acceptance tests.
//
// The container binds to a host port so that the Polytomic API server (running
// in Docker) can reach it via host.docker.internal.
func getSharedPGContainer(t *testing.T) pgContainer {
	t.Helper()

	sharedPGContainerOnce.Do(func() {
		ctx := context.Background()

		ctr, err := postgres.Run(ctx,
			"postgres:16-alpine",
			postgres.WithDatabase(testPGDatabase),
			postgres.WithUsername(testPGUser),
			postgres.WithPassword(testPGPassword),
			postgres.WithInitScripts(), // no file-based init scripts
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(30*time.Second),
			),
		)
		if err != nil {
			sharedPGContainerErr = fmt.Errorf("starting postgres container: %w", err)
			return
		}

		// Run init SQL to set up test schemas/tables.
		exitCode, _, execErr := ctr.Exec(ctx, []string{
			"psql", "-U", testPGUser, "-d", testPGDatabase, "-c", fixtureSQL,
		})
		if execErr != nil {
			sharedPGContainerErr = fmt.Errorf("running init SQL: %w", execErr)
			return
		}
		if exitCode != 0 {
			sharedPGContainerErr = fmt.Errorf("init SQL exited with code %d", exitCode)
			return
		}

		mappedPort, err := ctr.MappedPort(ctx, "5432")
		if err != nil {
			sharedPGContainerErr = fmt.Errorf("getting mapped port: %w", err)
			return
		}

		sharedPGContainer = &pgContainer{
			host: "host.docker.internal",
			port: int(mappedPort.Num()),
		}
	})

	if sharedPGContainerErr != nil {
		t.Fatalf("postgres testcontainer: %v", sharedPGContainerErr)
	}

	return *sharedPGContainer
}

// testPostgresConfig returns Postgres connection details for acceptance tests.
// It first checks for POLYTOMIC_TEST_PG_* environment variables and uses those
// if present. Otherwise it starts a shared Postgres testcontainer.
func testPostgresConfig(t *testing.T) postgresTestConfig {
	t.Helper()

	if cfg, ok := testPostgresConfigFromEnv(t); ok {
		return cfg
	}

	ctr := getSharedPGContainer(t)
	return postgresTestConfig{
		Host:     ctr.host,
		Database: testPGDatabase,
		Username: testPGUser,
		Password: testPGPassword,
		Port:     ctr.port,
	}
}
