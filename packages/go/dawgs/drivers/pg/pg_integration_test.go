package pg_test

import (
	"context"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/src/test/integration/utils"
	"github.com/stretchr/testify/require"
	"testing"
)

const murderDB = `
do
$$
declare
	t text;
begin
	for t in select relname from pg_class where oid in (select partrelid from pg_partitioned_table)
		loop
			execute 'drop table ' || t || ' cascade';
		end loop;

	for t in select table_name from information_schema.tables where table_schema = current_schema() and not table_name ilike '%pg_stat%'
		loop
			execute 'drop table ' || t || ' cascade';
		end loop;
end
$$;`

func Test_ResetDB(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	cfg, err := utils.LoadIntegrationTestConfig()
	require.Nil(t, err)

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		DriverCfg: cfg.Database.PostgreSQLConnectionString(),
	})
	require.Nil(t, err)

	require.Nil(t, graphDB.Run(ctx, murderDB, nil))
	require.Nil(t, graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema()))

	require.Nil(t, graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		user1, _ := tx.CreateNode(graph.AsProperties(map[string]any{
			"name": "user 1",
		}), ad.User)

		user2, _ := tx.CreateNode(graph.AsProperties(map[string]any{
			"name": "user 2",
		}), ad.User)

		group1, _ := tx.CreateNode(graph.AsProperties(map[string]any{
			"name": "group 1",
		}), ad.Group)

		computer1, _ := tx.CreateNode(graph.AsProperties(map[string]any{
			"name": "computer 1",
		}), ad.Computer)

		tx.CreateRelationshipByIDs(user1.ID, group1.ID, ad.MemberOf, nil)
		tx.CreateRelationshipByIDs(group1.ID, computer1.ID, ad.GenericAll, nil)
		tx.CreateRelationshipByIDs(computer1.ID, user2.ID, ad.HasSession, nil)

		return nil
	}))
}

func TestPG(t *testing.T) {
	ctx, done := context.WithCancel(context.Background())
	defer done()

	cfg, err := utils.LoadIntegrationTestConfig()
	require.Nil(t, err)

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		DriverCfg: cfg.Database.PostgreSQLConnectionString(),
	})
	require.Nil(t, err)

	require.Nil(t, graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Query("match p = (s:User)-[*..]->(:Computer) return p", nil).Error()
	}))
}
