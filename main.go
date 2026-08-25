package main

import (
	"fmt"
	"net/http"
	"os"

	"buckley-lev/internal/fluid"
	"buckley-lev/internal/report"
	"buckley-lev/internal/schema"
	"buckley-lev/internal/server"
	"buckley-lev/internal/welge"
)

const versionText = "buckley-lev 1.0.0"

const usageText = `buckley-lev —— 一维两相水驱 Buckley–Leverett 核算

用法:
  buckley-lev profile <算例.json> [--xi 0,0.3,1]   无因次剖面（ξ=x/t）与激波
  buckley-lev welge <算例.json>                    Welge 切点与激波运动学量
  buckley-lev fractional <算例.json> [--grid N]    分流函数 f(Sw) 采样表
  buckley-lev history <算例.json>                  突破前后出口/平均饱和度历史
  buckley-lev probe <算例.json>                    交叉规则实验（粘度/Sor/对称）
  buckley-lev sweep <算例.json> --param <p> --from <a> --to <b> [--n N]
                                                    单参数扫描（mu_w/mu_o/sor/...）
  buckley-lev check <算例.json>                    只做输入校验，不计算
  buckley-lev version                              打印版本
  buckley-lev help                                 显示本帮助

算例示例:
  buckley-lev profile example/waterflood.json
  buckley-lev sweep example/waterflood.json --param mu_w --from 0.1 --to 2 --n 8

说明:
  输入为 JSON（rock/relperm/fluid/injection），格式见 README 与
  example/waterflood.json。Swc/Sor/粘度/指数/注入饱和度非法、
  Welge 切线无解、Sw 越界都会报错并返回非零退出码。
`

type cliOptions struct {
	subcommand string
	file       string
	xi         string
	grid       int
	param      string
	from, to   float64
	steps      int
}

func parseArgs(args []string) (cliOptions, error) {
	var o cliOptions
	if len(args) == 0 {
		return o, fmt.Errorf("缺少子命令，运行 buckley-lev help 查看用法")
	}
	o.subcommand = args[0]
	o.grid = 0

	for i := 1; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--xi":
			if i+1 >= len(args) {
				return o, fmt.Errorf("--xi 需要一个逗号分隔的分数列表，如 0,0.5,1")
			}
			i++
			o.xi = args[i]
		case a == "--grid":
			if i+1 >= len(args) {
				return o, fmt.Errorf("--grid 需要一个正整数")
			}
			i++
			var n int
			if _, err := fmt.Sscanf(args[i], "%d", &n); err != nil || n < 3 || n > 1000 {
				return o, fmt.Errorf("--grid 必须是不小于 3 的整数，得到 %q", args[i])
			}
			o.grid = n
		case a == "--param":
			if i+1 >= len(args) {
				return o, fmt.Errorf("--param 需要一个参数名，如 mu_w / sor / krw0")
			}
			i++
			o.param = args[i]
		case a == "--from":
			if i+1 >= len(args) {
				return o, fmt.Errorf("--from 需要一个数值")
			}
			i++
			if _, err := fmt.Sscanf(args[i], "%g", &o.from); err != nil {
				return o, fmt.Errorf("--from 不是数值: %q", args[i])
			}
		case a == "--to":
			if i+1 >= len(args) {
				return o, fmt.Errorf("--to 需要一个数值")
			}
			i++
			if _, err := fmt.Sscanf(args[i], "%g", &o.to); err != nil {
				return o, fmt.Errorf("--to 不是数值: %q", args[i])
			}
		case a == "--n":
			if i+1 >= len(args) {
				return o, fmt.Errorf("--n 需要一个正整数")
			}
			i++
			var n int
			if _, err := fmt.Sscanf(args[i], "%d", &n); err != nil || n < 1 || n > 100 {
				return o, fmt.Errorf("--n 必须是 1~100 的整数，得到 %q", args[i])
			}
			o.steps = n
		case a == "-h" || a == "--help":
			o.subcommand = "help"
			return o, nil
		case a == "-v" || a == "--version":
			o.subcommand = "version"
			return o, nil
		case a == "":
			continue
		default:
			if o.file == "" {
				o.file = a
			} else {
				return o, fmt.Errorf("多余参数 %q", a)
			}
		}
	}
	return o, nil
}

func main() {
	args := os.Args[1:]
	if len(args) > 0 && (args[0] == "-http" || args[0] == "--http") {
		addr := ":8080"
		if len(args) > 1 {
			addr = args[1]
		}
		fmt.Printf("buckley-lev HTTP service on %s (POST /api/profile)\n", addr)
		if err := server.ListenAndServe(addr); err != nil && err != http.ErrServerClosed {
			fail("%v", err)
		}
		return
	}
	opts, err := parseArgs(args)
	if err != nil {
		fail("%v", err)
	}

	switch opts.subcommand {
	case "help", "-h", "--help":
		fmt.Print(usageText)

	case "version", "-v", "--version":
		fmt.Println(versionText)

	case "profile", "p":
		requireFile(opts, "profile")
		fractions, err := parseFractionsOpt(opts)
		if err != nil {
			fail("%v", err)
		}
		res, err := welge.SolveFile(opts.file, fractions)
		if err != nil {
			fail("%v", err)
		}
		fmt.Print(renderProfile(res))

	case "welge", "w":
		requireFile(opts, "welge")
		res, err := welge.SolveFile(opts.file, nil)
		if err != nil {
			fail("%v", err)
		}
		fmt.Print(report.RenderHeader(res))
		fmt.Print(report.RenderWelge(res))

	case "fractional", "f":
		requireFile(opts, "fractional")
		c, err := schema.ParseFile(opts.file)
		if err != nil {
			fail("%v", err)
		}
		schema.ApplyDefaults(c)
		if err := schema.Validate(c); err != nil {
			fail("%v", err)
		}
		m, err := fluid.NewModel(c.Rock.Swc, c.Rock.Sor,
			c.RelPerm.Krw0, c.RelPerm.Kro0,
			c.RelPerm.Nw, c.RelPerm.No,
			c.Fluid.MuW, c.Fluid.MuO)
		if err != nil {
			fail("%v", err)
		}
		grid := opts.grid
		if grid == 0 {
			grid = 21
		}
		fmt.Print(report.RenderFractional(m, grid))

	case "probe":
		requireFile(opts, "probe")
		c, err := schema.ParseFile(opts.file)
		if err != nil {
			fail("%v", err)
		}
		schema.ApplyDefaults(c)
		if err := schema.Validate(c); err != nil {
			fail("%v", err)
		}
		pr, err := welge.ProbeCrossRules(c)
		if err != nil {
			fail("%v", err)
		}
		fmt.Print(report.RenderProbe(pr))

	case "history":
		requireFile(opts, "history")
		res, err := welge.SolveFile(opts.file, nil)
		if err != nil {
			fail("%v", err)
		}
		an, err := welge.AnalyzePV(res.Model, res.Tangent, res.Case.Injection.SwInj, nil)
		if err != nil {
			fail("%v", err)
		}
		fmt.Print(report.RenderPVAnalysis(an))

	case "sweep":
		requireFile(opts, "sweep")
		c, err := schema.ParseFile(opts.file)
		if err != nil {
			fail("%v", err)
		}
		schema.ApplyDefaults(c)
		if err := schema.Validate(c); err != nil {
			fail("%v", err)
		}
		if opts.param == "" {
			fail("sweep 需要 --param，可用参数: mu_w mu_o sor swc krw0 kro0 nw no")
		}
		if opts.steps == 0 {
			opts.steps = 6
		}
		sr, err := welge.Sweep(c, welge.SweepParam(opts.param), opts.from, opts.to, opts.steps)
		if err != nil {
			fail("%v", err)
		}
		fmt.Print(report.RenderSweep(sr))

	case "check", "c":
		requireFile(opts, "check")
		c, err := schema.ParseFile(opts.file)
		if err != nil {
			fail("%v", err)
		}
		fmt.Print(report.RenderCheck(c))
		if err := schema.Validate(c); err != nil {
			os.Exit(1)
		}

	default:
		fail("未知子命令 %q，运行 buckley-lev help 查看用法", opts.subcommand)
	}
}

func renderProfile(res *welge.Result) string {
	out := ""
	out += report.RenderHeader(res) + "\n"
	out += report.RenderWelge(res) + "\n"
	out += report.RenderProfile(res) + "\n"
	out += report.SaturationChart(res, 60, 16)
	return out
}

func requireFile(o cliOptions, sub string) {
	if o.file == "" {
		fail("%s 需要一个算例文件参数，例如 example/waterflood.json", sub)
	}
}

func parseFractionsOpt(o cliOptions) ([]float64, error) {
	if o.xi == "" {
		return nil, nil
	}
	return welge.ParseFractions(o.xi)
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "buckley-lev: "+format+"\n", a...)
	os.Exit(1)
}
