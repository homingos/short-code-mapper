package redisStorage

const (
	prefix                  = "campaign-svc"
	universalCampaignPrefix = prefix + ":" + "universal-campaign"
	categoryPrefix          = prefix + ":" + "category"
)

func campaignExperiencesKey(campaignID string) string {
	return prefix + ":" + "campaign" + ":" + campaignID + ":" + "experiences"
}

func universalCampaignKey(universalCampaignID string) string {
	return universalCampaignPrefix + ":" + universalCampaignID + ":" + "experiences"
}

func categoryExperiencesKey(categoryID string) string {
	return categoryPrefix + ":" + categoryID + ":" + "experiences"
}
