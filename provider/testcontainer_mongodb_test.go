package provider

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testMongoUser     = "polytomic"
	testMongoPassword = "polytomic"
	testMongoDatabase = "polytomic"
)

// mongoFixtureJS seeds the collection the schema field tests override.
const mongoFixtureJS = `db.orders.insertOne({_id: "order-1", amount: "12.50", address: {city: "Paris"}})`

type mongoTestConfig struct {
	Hosts    string
	Database string
	Username string
	Password string
}

var (
	sharedMongoContainer     *mongoTestConfig
	sharedMongoContainerOnce sync.Once
	sharedMongoContainerErr  error
)

// testMongoConfig starts a MongoDB testcontainer once per test run and returns
// connection details for it. Like the Postgres container, it binds a host port
// so that the Polytomic API server (running in Docker) can reach it via
// host.docker.internal.
func testMongoConfig(t *testing.T) mongoTestConfig {
	t.Helper()

	sharedMongoContainerOnce.Do(func() {
		ctx := context.Background()

		ctr, err := testcontainers.Run(ctx, "mongo:7",
			testcontainers.WithExposedPorts("27017/tcp"),
			testcontainers.WithEnv(map[string]string{
				"MONGO_INITDB_ROOT_USERNAME": testMongoUser,
				"MONGO_INITDB_ROOT_PASSWORD": testMongoPassword,
			}),
			// The image restarts mongod after creating the root user, so the
			// second "Waiting for connections" is the one that accepts auth.
			testcontainers.WithWaitStrategy(wait.ForAll(
				wait.ForLog("Waiting for connections").WithOccurrence(2),
				wait.ForListeningPort("27017/tcp"),
			).WithDeadline(90*time.Second)),
		)
		if err != nil {
			sharedMongoContainerErr = fmt.Errorf("starting mongo container: %w", err)
			return
		}

		exitCode, _, err := ctr.Exec(ctx, []string{
			"mongosh", "--quiet",
			"--username", testMongoUser,
			"--password", testMongoPassword,
			"--authenticationDatabase", "admin",
			testMongoDatabase,
			"--eval", mongoFixtureJS,
		})
		if err != nil {
			sharedMongoContainerErr = fmt.Errorf("seeding mongo: %w", err)
			return
		}
		if exitCode != 0 {
			sharedMongoContainerErr = fmt.Errorf("seeding mongo exited with code %d", exitCode)
			return
		}

		mappedPort, err := ctr.MappedPort(ctx, "27017/tcp")
		if err != nil {
			sharedMongoContainerErr = fmt.Errorf("getting mapped port: %w", err)
			return
		}

		sharedMongoContainer = &mongoTestConfig{
			Hosts:    fmt.Sprintf("host.docker.internal:%d", mappedPort.Num()),
			Database: testMongoDatabase,
			Username: testMongoUser,
			Password: testMongoPassword,
		}
	})

	if sharedMongoContainerErr != nil {
		t.Fatalf("mongo testcontainer: %v", sharedMongoContainerErr)
	}

	return *sharedMongoContainer
}
