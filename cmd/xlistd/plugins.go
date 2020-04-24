// Copyright 2019 Luis Guillén Civera <luisguillenc@gmail.com>. View LICENSE.

package main

import (
	//components
	_ "github.com/luids-io/xlist/pkg/components/dnsxl"
	_ "github.com/luids-io/xlist/pkg/components/filexl"
	_ "github.com/luids-io/xlist/pkg/components/geoip2xl"
	_ "github.com/luids-io/xlist/pkg/components/grpcxl"
	_ "github.com/luids-io/xlist/pkg/components/memxl"
	_ "github.com/luids-io/xlist/pkg/components/mockxl"
	_ "github.com/luids-io/xlist/pkg/components/parallelxl"
	_ "github.com/luids-io/xlist/pkg/components/selectorxl"
	_ "github.com/luids-io/xlist/pkg/components/sequencexl"
	_ "github.com/luids-io/xlist/pkg/components/wbeforexl"

	//wrappers
	_ "github.com/luids-io/xlist/pkg/wrappers/cachewr"
	_ "github.com/luids-io/xlist/pkg/wrappers/loggerwr"
	_ "github.com/luids-io/xlist/pkg/wrappers/metricswr"
	_ "github.com/luids-io/xlist/pkg/wrappers/policywr"
	_ "github.com/luids-io/xlist/pkg/wrappers/responsewr"
	_ "github.com/luids-io/xlist/pkg/wrappers/scorewr"
	_ "github.com/luids-io/xlist/pkg/wrappers/timeoutwr"
)
