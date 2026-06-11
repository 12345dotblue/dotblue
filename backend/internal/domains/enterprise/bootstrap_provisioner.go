package enterprise

import (
	"context"
	"fmt"

	"dotblue/internal/domains/credit"
	"dotblue/internal/domains/settings"
	"github.com/gogf/gf/v2/frame/g"
)

const enterpriseBootstrapGrantSourceType = "enterprise_bootstrap"

type BootstrapProvisioner interface {
	BootstrapNewEnterprise(enterpriseId string) error
}

type platformBootstrapProvisioner struct{}
type noopBootstrapProvisioner struct{}

func newPlatformBootstrapProvisioner() BootstrapProvisioner {
	return &platformBootstrapProvisioner{}
}

func (noopBootstrapProvisioner) BootstrapNewEnterprise(string) error {
	return nil
}

func (p *platformBootstrapProvisioner) BootstrapNewEnterprise(enterpriseId string) error {
	cfg, err := settings.GetPlatformConfig()
	if err != nil {
		return err
	}
	if cfg == nil || cfg.NewEnterprisePlatformCredits <= 0 {
		return nil
	}
	metadata := fmt.Sprintf(`{"source":"platform_config","configuredCredits":%d}`, cfg.NewEnterprisePlatformCredits)
	grant, _, err := credit.CreateGrant(enterpriseId, credit.CreditTypePlatform, credit.CreateGrantReq{
		SourceType:   enterpriseBootstrapGrantSourceType,
		SourceRefId:  enterpriseId,
		Credits:      cfg.NewEnterprisePlatformCredits,
		MetadataJson: metadata,
		ReasonCode:   enterpriseBootstrapGrantSourceType,
	})
	if err != nil {
		g.Log().Warningf(context.Background(), "enterprise.bootstrap_credits.failed enterprise=%s err=%v", enterpriseId, err)
		return err
	}
	if grant != nil {
		g.Log().Infof(context.Background(), "enterprise.bootstrap_credits.applied enterprise=%s credits=%d grant=%s", enterpriseId, grant.GrantedCredits, grant.Id)
	}
	return nil
}
