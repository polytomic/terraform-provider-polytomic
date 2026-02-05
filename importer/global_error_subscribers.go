package importer

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	ptclient "github.com/polytomic/polytomic-go/client"
)

const (
	GlobalErrorSubscribersResourceFileName = "global_error_subscribers.tf"
	GlobalErrorSubscribersResourceType     = "polytomic_global_error_subscribers"
	globalErrorSubscribersImportID         = "global-error-subscribers"
)

var _ Importable = &GlobalErrorSubscribers{}

type GlobalErrorSubscribers struct {
	c      *ptclient.Client
	emails []string
}

func NewGlobalErrorSubscribers(c *ptclient.Client) *GlobalErrorSubscribers {
	return &GlobalErrorSubscribers{c: c}
}

func (g *GlobalErrorSubscribers) Init(ctx context.Context) error {
	resp, err := g.c.Notifications.GetGlobalErrorSubscribers(ctx)
	if err != nil {
		return err
	}
	if resp == nil {
		return fmt.Errorf("empty response")
	}

	g.emails = resp.Emails
	if g.emails == nil {
		g.emails = []string{}
	}

	return nil
}

func (g *GlobalErrorSubscribers) GenerateTerraformFiles(ctx context.Context, writer io.Writer, refs map[string]string) error {
	// If nothing is configured, don't generate a resource block.
	if len(g.emails) == 0 {
		return nil
	}

	hclFile := hclwrite.NewEmptyFile()
	body := hclFile.Body()

	resourceBlock := body.AppendNewBlock("resource", []string{GlobalErrorSubscribersResourceType, "global"})
	resourceBlock.Body().SetAttributeTraversal("organization",
		hcl.Traversal{
			hcl.TraverseRoot{Name: "local"},
			hcl.TraverseAttr{Name: "organization_id"},
		},
	)
	resourceBlock.Body().SetAttributeValue("emails", typeConverter(g.emails))
	body.AppendNewline()

	_, err := writer.Write(hclFile.Bytes())
	return err
}

func (g *GlobalErrorSubscribers) GenerateImports(ctx context.Context, writer io.Writer) error {
	if len(g.emails) == 0 {
		return nil
	}
	_, err := fmt.Fprintf(writer, "terraform import %s.%s %s\n",
		GlobalErrorSubscribersResourceType,
		"global",
		globalErrorSubscribersImportID,
	)
	return err
}

func (g *GlobalErrorSubscribers) Filename() string {
	return GlobalErrorSubscribersResourceFileName
}

func (g *GlobalErrorSubscribers) ResourceRefs() map[string]string {
	return nil
}

func (g *GlobalErrorSubscribers) DatasourceRefs() map[string]string {
	return nil
}

func (g *GlobalErrorSubscribers) Variables() []Variable {
	return nil
}
