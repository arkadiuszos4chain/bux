package chainstate

import (
	"context"

	"github.com/mrz1836/go-mattercloud"
	"github.com/mrz1836/go-nownodes"
	"github.com/mrz1836/go-whatsonchain"
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/tonicpow/go-minercraft"
)

// defaultMinercraftOptions will create the defaults
func (c *Client) defaultMinercraftOptions() (opts *minercraft.ClientOptions) {
	opts = minercraft.DefaultClientOptions()
	if len(c.options.userAgent) > 0 {
		opts.UserAgent = c.options.userAgent
	}
	return
}

// defaultWhatsOnChainOptions will create the defaults
func (c *Client) defaultWhatsOnChainOptions() (opts *whatsonchain.Options) {
	opts = whatsonchain.ClientDefaultOptions()
	if len(c.options.userAgent) > 0 {
		opts.UserAgent = c.options.userAgent
	}

	// Set a custom API key
	// todo: rate limit should be customizable
	if len(c.options.config.whatsOnChainAPIKey) > 0 {
		opts.APIKey = c.options.config.whatsOnChainAPIKey
		opts.RateLimit = whatsOnChainRateLimitWithKey
	}
	return
}

// defaultMatterCloudOptions will create the defaults
func (c *Client) defaultMatterCloudOptions() (opts *mattercloud.Options) {
	opts = mattercloud.ClientDefaultOptions()
	if len(c.options.userAgent) > 0 {
		opts.UserAgent = c.options.userAgent
	}
	return
}

// startMinerCraft will start Minercraft (if no custom client is found)
func (c *Client) startMinerCraft(ctx context.Context) (err error) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("start_minercraft").End()
	}

	// No client?
	if c.Minercraft() == nil {
		var optionalMiners []*minercraft.Miner
		for i := range c.options.config.mAPI.miners {
			optionalMiners = append(optionalMiners, c.options.config.mAPI.miners[i].Miner)
		}
		c.options.config.minercraft, err = minercraft.NewClient(
			c.defaultMinercraftOptions(),
			c.HTTPClient(),
			optionalMiners,
		)
	}

	// Check for broadcast miners
	if len(c.BroadcastMiners()) == 0 {
		return ErrMissingBroadcastMiners
	}

	// Check for query miners
	if len(c.QueryMiners()) == 0 {
		return ErrMissingQueryMiners
	}

	// Load the quote fees
	return c.RefreshFeeQuotes(ctx)
}

// startWhatsOnChain will start WhatsOnChain (if no custom client is found)
func (c *Client) startWhatsOnChain(ctx context.Context) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("start_whatsonchain").End()
	}

	if c.WhatsOnChain() == nil {
		c.options.config.whatsOnChain = whatsonchain.NewClient(
			c.Network().WhatsOnChain(),
			c.defaultWhatsOnChainOptions(),
			c.HTTPClient(),
		)
	}
}

// startMatterCloud will start MatterCloud (if no custom client is found)
func (c *Client) startMatterCloud(ctx context.Context) (err error) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("start_mattercloud").End()
	}

	if c.MatterCloud() == nil {

		// go-matter cloud requires a key to load the library
		if len(c.options.config.matterCloudAPIKey) == 0 {
			c.options.config.matterCloudAPIKey = "your-api-key" // todo: hack for go-mattercloud (wants an api key)
		}

		c.options.config.matterCloud, err = mattercloud.NewClient(
			c.options.config.matterCloudAPIKey,
			c.Network().MatterCloud(),
			c.defaultMatterCloudOptions(),
			c.HTTPClient(),
		)
	}
	return
}

// startNowNodes will start NowNodes if API key is set (if no custom client is found)
func (c *Client) startNowNodes(ctx context.Context) {
	if txn := newrelic.FromContext(ctx); txn != nil {
		defer txn.StartSegment("start_nownodes").End()
	}

	if c.NowNodes() == nil && len(c.options.config.nowNodesAPIKey) > 0 {
		c.options.config.nowNodes = nownodes.NewClient(
			nownodes.WithAPIKey(c.options.config.nowNodesAPIKey),
			nownodes.WithHTTPClient(c.HTTPClient()),
		)
	}
}
