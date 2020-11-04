// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

import (
	//components
	_ "github.com/luids-io/xlist/pkg/xlistd/components/apicheckxl"
	_ "github.com/luids-io/xlist/pkg/xlistd/components/dnsxl"
	_ "github.com/luids-io/xlist/pkg/xlistd/components/filexl"
	_ "github.com/luids-io/xlist/pkg/xlistd/components/geoip2xl"
	_ "github.com/luids-io/xlist/pkg/xlistd/components/grpcxl"
	_ "github.com/luids-io/xlist/pkg/xlistd/components/memxl"
	_ "github.com/luids-io/xlist/pkg/xlistd/components/mockxl"
	_ "github.com/luids-io/xlist/pkg/xlistd/components/parallelxl"
	_ "github.com/luids-io/xlist/pkg/xlistd/components/sblookupxl"
	_ "github.com/luids-io/xlist/pkg/xlistd/components/selectorxl"
	_ "github.com/luids-io/xlist/pkg/xlistd/components/sequencexl"
	_ "github.com/luids-io/xlist/pkg/xlistd/components/wbeforexl"

	//wrappers
	_ "github.com/luids-io/xlist/pkg/xlistd/wrappers/cachewr"
	_ "github.com/luids-io/xlist/pkg/xlistd/wrappers/loggerwr"
	_ "github.com/luids-io/xlist/pkg/xlistd/wrappers/metricswr"
	_ "github.com/luids-io/xlist/pkg/xlistd/wrappers/policywr"
	_ "github.com/luids-io/xlist/pkg/xlistd/wrappers/responsewr"
	_ "github.com/luids-io/xlist/pkg/xlistd/wrappers/scorewr"
	_ "github.com/luids-io/xlist/pkg/xlistd/wrappers/timeoutwr"
)
