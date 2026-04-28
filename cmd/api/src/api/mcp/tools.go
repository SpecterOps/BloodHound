package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	mcpsdk "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAllTools(srv *server.MCPServer, loopbackURL string) {
	registerCypherQuery(srv, loopbackURL)
	registerSearch(srv, loopbackURL)
	registerShortestPath(srv, loopbackURL)
	registerListDomains(srv, loopbackURL)
	registerGetEntity(srv, loopbackURL)
	registerListNodeKinds(srv, loopbackURL)
	registerMarkOwned(srv, loopbackURL)
	registerAddTag(srv, loopbackURL)
	registerListTags(srv, loopbackURL)
}

func clientFromRequest(req mcpsdk.CallToolRequest, loopbackURL string) (*client, error) {
	header := req.Header.Get("X-BH-Token")
	creds, ok := tokenFromHeader(header)
	if !ok {
		return nil, fmt.Errorf("missing or invalid X-BH-Token header")
	}
	return newClient(loopbackURL, creds.ID, creds.Key), nil
}

// --- Read Tools ---

func registerCypherQuery(srv *server.MCPServer, loopbackURL string) {
	tool := mcpsdk.NewTool("cypher_query",
		mcpsdk.WithDescription("Execute a Cypher query against the BloodHound graph database. Must return graph objects (nodes/edges), not just properties. Examples: MATCH (n:User) RETURN n LIMIT 10, MATCH p=shortestPath((a)-[*1..]->(b:Domain)) RETURN p"),
		mcpsdk.WithString("query", mcpsdk.Required(), mcpsdk.Description("The Cypher query to execute")),
		mcpsdk.WithBoolean("include_properties", mcpsdk.Description("Include node/edge properties in response (default: true)")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		c, err := clientFromRequest(req, loopbackURL)
		if err != nil {
			return mcpsdk.NewToolResultError(err.Error()), nil
		}
		query, err := req.RequireString("query")
		if err != nil {
			return mcpsdk.NewToolResultError("query parameter is required"), nil
		}
		body := map[string]any{
			"query":              query,
			"include_properties": req.GetBool("include_properties", true),
		}
		data, err := c.Post(ctx, "/api/v2/graphs/cypher", body)
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("Cypher query failed: %v", err)), nil
		}
		return mcpsdk.NewToolResultText(string(data)), nil
	})
}

func registerSearch(srv *server.MCPServer, loopbackURL string) {
	tool := mcpsdk.NewTool("search",
		mcpsdk.WithDescription("Search for Active Directory or Azure objects by name or object ID."),
		mcpsdk.WithString("query", mcpsdk.Required(), mcpsdk.Description("Search keyword (name or object ID)")),
		mcpsdk.WithString("type", mcpsdk.Description("Filter by node type (e.g. User, Computer, Group, Domain)")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		c, err := clientFromRequest(req, loopbackURL)
		if err != nil {
			return mcpsdk.NewToolResultError(err.Error()), nil
		}
		query, err := req.RequireString("query")
		if err != nil {
			return mcpsdk.NewToolResultError("query parameter is required"), nil
		}
		params := url.Values{"q": {query}}
		if nodeType := req.GetString("type", ""); nodeType != "" {
			params.Set("type", nodeType)
		}
		data, err := c.Get(ctx, "/api/v2/search", params)
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
		}
		return mcpsdk.NewToolResultText(string(data)), nil
	})
}

func registerShortestPath(srv *server.MCPServer, loopbackURL string) {
	tool := mcpsdk.NewTool("shortest_path",
		mcpsdk.WithDescription("Find the shortest attack path between two nodes. Use the search tool first to find object IDs."),
		mcpsdk.WithString("start_node", mcpsdk.Required(), mcpsdk.Description("Object ID of the start node")),
		mcpsdk.WithString("end_node", mcpsdk.Required(), mcpsdk.Description("Object ID of the end node")),
		mcpsdk.WithString("relationship_kinds", mcpsdk.Description("Filter by relationship types, format: 'in:Kind1,Kind2' or 'nin:Kind1,Kind2'")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		c, err := clientFromRequest(req, loopbackURL)
		if err != nil {
			return mcpsdk.NewToolResultError(err.Error()), nil
		}
		startNode, err := req.RequireString("start_node")
		if err != nil {
			return mcpsdk.NewToolResultError("start_node is required"), nil
		}
		endNode, err := req.RequireString("end_node")
		if err != nil {
			return mcpsdk.NewToolResultError("end_node is required"), nil
		}
		params := url.Values{"start_node": {startNode}, "end_node": {endNode}}
		if kinds := req.GetString("relationship_kinds", ""); kinds != "" {
			params.Set("relationship_kinds", kinds)
		}
		data, err := c.Get(ctx, "/api/v2/graphs/shortest-path", params)
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("Shortest path failed: %v", err)), nil
		}
		return mcpsdk.NewToolResultText(string(data)), nil
	})
}

func registerListDomains(srv *server.MCPServer, loopbackURL string) {
	tool := mcpsdk.NewTool("list_domains",
		mcpsdk.WithDescription("List all available Active Directory domains and Azure tenants."),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		c, err := clientFromRequest(req, loopbackURL)
		if err != nil {
			return mcpsdk.NewToolResultError(err.Error()), nil
		}
		data, err := c.Get(ctx, "/api/v2/available-domains", nil)
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("List domains failed: %v", err)), nil
		}
		return mcpsdk.NewToolResultText(string(data)), nil
	})
}

var entityTypePaths = map[string]string{
	"user": "users", "computer": "computers", "group": "groups",
	"domain": "domains", "ou": "ous", "gpo": "gpos", "container": "containers",
	"aiaca": "aiacas", "rootca": "rootcas", "enterpriseca": "enterprisecas",
	"ntauthstore": "ntauthstores", "certtemplate": "certtemplates",
	"issuancepolicy": "issuancepolicies", "base": "base",
}

func registerGetEntity(srv *server.MCPServer, loopbackURL string) {
	tool := mcpsdk.NewTool("get_entity",
		mcpsdk.WithDescription("Get detailed info about a specific AD/Azure entity. Use search to find object IDs first."),
		mcpsdk.WithString("object_id", mcpsdk.Required(), mcpsdk.Description("The object ID of the entity")),
		mcpsdk.WithString("entity_type", mcpsdk.Required(), mcpsdk.Description("Entity type: user, computer, group, domain, ou, gpo, container, aiaca, rootca, enterpriseca, ntauthstore, certtemplate, issuancepolicy, or base")),
		mcpsdk.WithBoolean("counts", mcpsdk.Description("Include relationship counts (default: true)")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		c, err := clientFromRequest(req, loopbackURL)
		if err != nil {
			return mcpsdk.NewToolResultError(err.Error()), nil
		}
		objectID, err := req.RequireString("object_id")
		if err != nil {
			return mcpsdk.NewToolResultError("object_id is required"), nil
		}
		entityType, err := req.RequireString("entity_type")
		if err != nil {
			return mcpsdk.NewToolResultError("entity_type is required"), nil
		}
		pathSeg, ok := entityTypePaths[entityType]
		if !ok {
			return mcpsdk.NewToolResultError(fmt.Sprintf("unknown entity_type %q", entityType)), nil
		}
		params := url.Values{}
		if req.GetBool("counts", true) {
			params.Set("counts", "true")
		}
		data, err := c.Get(ctx, fmt.Sprintf("/api/v2/%s/%s", pathSeg, objectID), params)
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("Get entity failed: %v", err)), nil
		}
		return mcpsdk.NewToolResultText(string(data)), nil
	})
}

func registerListNodeKinds(srv *server.MCPServer, loopbackURL string) {
	tool := mcpsdk.NewTool("list_node_kinds",
		mcpsdk.WithDescription("List all node and relationship kinds in the BloodHound graph schema."),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		c, err := clientFromRequest(req, loopbackURL)
		if err != nil {
			return mcpsdk.NewToolResultError(err.Error()), nil
		}
		data, err := c.Get(ctx, "/api/v2/graphs/kinds", nil)
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("List node kinds failed: %v", err)), nil
		}
		return mcpsdk.NewToolResultText(string(data)), nil
	})
}

// --- Write Tools ---

func registerMarkOwned(srv *server.MCPServer, loopbackURL string) {
	tool := mcpsdk.NewTool("mark_owned",
		mcpsdk.WithDescription("Mark an entity as owned/compromised in BloodHound. Creates or reuses an 'Owned' asset group tag and adds the entity to it."),
		mcpsdk.WithString("object_id", mcpsdk.Required(), mcpsdk.Description("Object ID of the compromised entity")),
		mcpsdk.WithString("tag_name", mcpsdk.Description("Name for the owned tag (default: 'Owned - MCP')")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		c, err := clientFromRequest(req, loopbackURL)
		if err != nil {
			return mcpsdk.NewToolResultError(err.Error()), nil
		}
		objectID, err := req.RequireString("object_id")
		if err != nil {
			return mcpsdk.NewToolResultError("object_id is required"), nil
		}
		tagName := req.GetString("tag_name", "Owned - MCP")

		// Find existing owned tag or create one
		tagID, err := findOrCreateOwnedTag(ctx, c, tagName)
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("Failed to get/create owned tag: %v", err)), nil
		}

		// Add entity as a selector seed
		selectorBody := map[string]any{
			"name": fmt.Sprintf("owned-%s", objectID),
			"seeds": []map[string]any{
				{"type": 1, "value": objectID},
			},
			"auto_certify": 0,
			"disabled":     false,
		}
		_, err = c.Post(ctx, fmt.Sprintf("/api/v2/asset-group-tags/%d/selectors", tagID), selectorBody)
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("Failed to add entity to owned tag: %v", err)), nil
		}

		return mcpsdk.NewToolResultText(fmt.Sprintf("Entity %s marked as owned (tag: %s, id: %d)", objectID, tagName, tagID)), nil
	})
}

func findOrCreateOwnedTag(ctx context.Context, c *client, tagName string) (int, error) {
	// List existing tags, look for matching owned tag
	data, err := c.Get(ctx, "/api/v2/asset-group-tags", url.Values{"counts": {"true"}})
	if err != nil {
		return 0, fmt.Errorf("listing tags: %w", err)
	}

	var tags []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Type int    `json:"tag_type"`
	}
	if err := json.Unmarshal(data, &tags); err != nil {
		// Try unwrapping from {"asset_group_tags": [...]}
		var wrapper struct {
			Tags json.RawMessage `json:"asset_group_tags"`
		}
		if err2 := json.Unmarshal(data, &wrapper); err2 == nil && wrapper.Tags != nil {
			json.Unmarshal(wrapper.Tags, &tags)
		}
	}

	for _, t := range tags {
		if t.Type == 3 && t.Name == tagName {
			return t.ID, nil
		}
	}

	// Create new owned tag (type=3, glyph=skull)
	createBody := map[string]any{
		"name":  tagName,
		"type":  3,
		"glyph": "skull",
	}
	respData, err := c.Post(ctx, "/api/v2/asset-group-tags", createBody)
	if err != nil {
		return 0, fmt.Errorf("creating tag: %w", err)
	}

	var created struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(respData, &created); err != nil {
		return 0, fmt.Errorf("parsing created tag: %w", err)
	}
	return created.ID, nil
}

func registerAddTag(srv *server.MCPServer, loopbackURL string) {
	tool := mcpsdk.NewTool("add_tag",
		mcpsdk.WithDescription("Add an entity to an existing asset group tag by tag ID. Use list_tags to find tag IDs."),
		mcpsdk.WithString("object_id", mcpsdk.Required(), mcpsdk.Description("Object ID of the entity to tag")),
		mcpsdk.WithNumber("tag_id", mcpsdk.Required(), mcpsdk.Description("Asset group tag ID to add the entity to")),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		c, err := clientFromRequest(req, loopbackURL)
		if err != nil {
			return mcpsdk.NewToolResultError(err.Error()), nil
		}
		objectID, err := req.RequireString("object_id")
		if err != nil {
			return mcpsdk.NewToolResultError("object_id is required"), nil
		}
		tagID := req.GetInt("tag_id", 0)
		if tagID == 0 {
			return mcpsdk.NewToolResultError("tag_id is required"), nil
		}

		selectorBody := map[string]any{
			"name": fmt.Sprintf("mcp-%s", objectID),
			"seeds": []map[string]any{
				{"type": 1, "value": objectID},
			},
			"auto_certify": 0,
			"disabled":     false,
		}
		_, err = c.Post(ctx, fmt.Sprintf("/api/v2/asset-group-tags/%d/selectors", tagID), selectorBody)
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("Failed to add entity to tag: %v", err)), nil
		}
		return mcpsdk.NewToolResultText(fmt.Sprintf("Entity %s added to tag %d", objectID, tagID)), nil
	})
}

func registerListTags(srv *server.MCPServer, loopbackURL string) {
	tool := mcpsdk.NewTool("list_tags",
		mcpsdk.WithDescription("List all asset group tags (tiers, labels, owned) with member counts."),
	)
	srv.AddTool(tool, func(ctx context.Context, req mcpsdk.CallToolRequest) (*mcpsdk.CallToolResult, error) {
		c, err := clientFromRequest(req, loopbackURL)
		if err != nil {
			return mcpsdk.NewToolResultError(err.Error()), nil
		}
		data, err := c.Get(ctx, "/api/v2/asset-group-tags", url.Values{"counts": {"true"}})
		if err != nil {
			return mcpsdk.NewToolResultError(fmt.Sprintf("List tags failed: %v", err)), nil
		}
		return mcpsdk.NewToolResultText(string(data)), nil
	})
}
