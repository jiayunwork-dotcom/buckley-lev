# buckley-lev

buckley-lev 是一维两相水驱（Buckley–Leverett）核算命令行工具：用户以 JSON 给出油水相对渗透率（Corey 幂次）、粘度与注入饱和度，工具计算分流函数 f(Sw)、Welge 切点饱和度 Swf、激波速度与激波后方由 df/dSw 决定的稀疏波饱和度剖面（无因次速度 ξ=x/t）。

- 输入：`example/waterflood.json` 格式的算例文件（rock/relperm/fluid/injection 四组字段）
- 输出：Swf、无因次激波速度 ξs、突破注入孔隙体积、若干 ξ 处的 Sw、ASCII 剖面图、Welge 物质平衡（出口/平均饱和度随 PV 的历史）、交叉规则探针
- 模型：Corey 相对渗透率（krw=krw0·S*ⁿʷ，kro=kro0·(1−S*)^ⁿᵒ），分流函数 f(Sw)=λw/(λw+λo)；前缘由 Welge 切线（从 (Swc,0) 作割线）定位，激波满足 Rankine–Hugoniot，激波后方用稀疏波反解而不是一条斜直线
- 边界：一维、水油两相、无压缩性；不做化学驱、不做网格离散、不做原油采购或驱替工单

## 快速开始

```text
go run . profile example/waterflood.json
```

输出示例（摘要）：

```text
前缘饱和度 Swf        = 0.636564
切点局部斜率 f'(Swf)  = 1.978640
割线斜率 f(Swf)/(Swf−Swc) = 1.978640
无因次激波速度 ξs     = 1.978640
突破注入孔隙体积      = 0.5054 PV
```

## 子命令

| 子命令 | 说明 |
| --- | --- |
| `profile <算例.json> [--xi 0,0.3,1]` | 无因次剖面（ξ=x/t）与激波（主命令） |
| `welge <算例.json>` | Welge 切点与激波运动学量 |
| `fractional <算例.json> [--grid N]` | 分流函数 f(Sw) 采样表 |
| `history <算例.json>` | 突破前后出口/平均饱和度与采收率随 PV 的历史 |
| `probe <算例.json>` | 交叉规则实验（粘度/Sor/对称） |
| `sweep <算例.json> --param <p> --from <a> --to <b> [--n N]` | 单参数扫描（mu_w/mu_o/sor/swc/krw0/kro0/nw/no） |
| `check <算例.json>` | 只校验输入，不计算 |
| `help` | 显示帮助 |

非法输入（Swc<0、Sor<0、Swc+Sor≥1、粘度≤0、Corey 指数≤1、注入饱和度越界、JSON 错误、Welge 切线无解、Sw 越界）一律写入 stderr 并返回非零退出码。

## 关键约定

- **归一化饱和度**：S* = (Sw−Swc)/(1−Swc−Sor)，端点处 f(Swc)=0、f(1−Sor)=1
- **分流函数**：f(Sw) = λw/(λw+λo)，λ=kr/μ；f 是 S 形单调函数，与 krw 本身严格区分
- **Welge 切点**：Swf 由从 (Swc,0) 向 f(Sw) 作切线确定，切点处 f'(Swf)=f(Swf)/(Swf−Swc)
- **激波**：无因次速度 ξs = f(Swf)/(Swf−Swc)，即 Rankine–Hugoniot 割线斜率；突破注入孔隙体积 = 1/ξs
- **稀疏波**：激波后方 0<ξ<ξs 处饱和度由 f'(Sw)=ξ 反解，是特征线解，不是线性插值
- **交叉规则**（物理方向）：只降低水粘度 → 水相流度上升、Swf 下降、前缘变尖、激波速度上升；只加大 Sor → 可动油变少、末端饱和度 1−Sor 左移；μw=μo 且 krw0=kro0、指数相同 → f 关于可动油带中点对称、f(中点)=0.5
- **指数约束**：Corey 指数必须 > 1，线性相对渗透率（n=1）会使切线构造退化

## 算例文件格式

```json
{
  "rock": { "swc": 0.2, "sor": 0.2, "porosity": 0.25 },
  "relperm": { "krw0": 0.4, "kro0": 0.9, "nw": 2.0, "no": 2.0 },
  "fluid": { "mu_w": 1.0, "mu_o": 2.0 },
  "injection": { "sw_inj": 0.8 }
}
```

`rock.porosity` 可省略（默认 0.25，只用于突破孔隙体积换算）；其余字段必须给出。`sw_inj` 一般取 1−Sor（注入水把可动油全部排出）。

## 构建 / 测试

```text
go build ./...   # 纯标准库，无第三方依赖
go test ./...    # schema / fluid / welge / report 四个包
go vet ./...
```

## 目录

```text
main.go                    CLI 子命令接线
internal/schema/           输入模型、严格 JSON 解析、默认值与物理校验
internal/fluid/            Corey 相对渗透率、分流函数与解析导数、单调/对称检查
internal/welge/            Welge 切点、激波、稀疏波反解、剖面、物质平衡、扫描
internal/report/           文本表格、ASCII 剖面图、各子命令渲染
example/waterflood.json    水驱算例
```

## Docker

```text
docker build -t buckley-lev .
docker run --rm -it buckley-lev profile example/waterflood.json
```

## License

MIT。
