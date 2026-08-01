package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

type ClashConfig struct {
	Proxies []map[string]any `yaml:"proxies"`
	// 可添加 proxy-groups, rules 等
	ProxyGroups []ProxyGroup `yaml:"proxy-groups"`
}

// 分组配置结构体
type ProxyGroup struct {
	Name                string    `yaml:"name"` // 必选字段
	Type                string    `yaml:"type"` // 必选字段
	Use                 *[]string `yaml:"use,omitempty"`
	URL                 string    `yaml:"url,omitempty"`
	Interval            int       `yaml:"interval,omitempty"`
	Lazy                bool      `yaml:"lazy,omitempty"`
	Timeout             int       `yaml:"timeout,omitempty"`
	MaxFailedTimes      int       `yaml:"max-failed-times,omitempty"`
	DisableUDP          bool      `yaml:"disable-udp,omitempty"`
	InterfaceName       *string   `yaml:"interface-name,omitempty"`
	RoutingMark         *int      `yaml:"routing-mark,omitempty"`
	IncludeAll          bool      `yaml:"include-all,omitempty"`
	IncludeAllProxies   *bool     `yaml:"include-all-proxies,omitempty"`
	IncludeAllProviders *bool     `yaml:"include-all-providers,omitempty"`
	Filter              string    `yaml:"filter,omitempty"`
	ExcludeFilter       *string   `yaml:" exclude-filter,omitempty"`
	ExcludeType         *string   `yaml:" exclude-type,omitempty"` // "Shadowsocks|Http"
	ExpectedStatus      *int      `yaml:"exclude-type,omitempty"`
	Hidden              bool      `yaml:"hidden,omitempty"`
	Proxies             []string  `yaml:"proxies,omitempty"`
	Icon                string    `yaml:"icon,omitempty"`
}

// 添加一个辅助函数用于创建指针值
func ptr[T any](v T) *T {
	return &v
}

var (
	// FilterNotChina = &yaml.Node{
	// 	Kind:   yaml.ScalarNode,
	// 	Tag:    "!!str",
	// 	Value:  "^(?!.*🇨🇳)(?!.*中国).*",
	// 	Anchor: "FilterNotChina",
	// }

	// 定义你的 key 和正则值（含锚点）
	entries = map[string]string{
		// "FilterFR":       "^(?=.*((?i)🇫🇷|法国|(\\b(FR|France)\\b))).*$",
		// "FilterDE":       "^(?=.*((?i)🇩🇪|德国|(\\b(DE|Germany)\\b))).*$",
		// "FilterBR":       "^(?=.*((?i)🇧🇷|巴西|\\b(BR|Brazil)\\b)).*$",
		// "FilterPE":       "^(?=.*((?i)🇵🇪|秘鲁|\\b(PE|Peru)\\b)).*$",
		// "FilterZA":       "^(?=.*((?i)🇿🇦|南非|\\b(ZA|South Africa)\\b)).*$",
		// "FilterCZ":       "^(?=.*((?i)🇨🇿|捷克|\\b(CZ|Czech)\\b)).*$",
		// "FilterNL":       "^(?=.*((?i)🇳🇱|荷兰|\\b(NL|Netherlands)\\b)).*$",
		// "FilterTR":       "^(?=.*((?i)🇹🇷|土耳其|\\b(TR|Turkey)\\b)).*$",
		"FilterGame":     "^(?=.*((?i)游戏|🎮|(\b(GAME)\b)))(?!.*((?i)回国|校园)).*$",
		"FilterNotChina": "^(?!.*🇨🇳)(?!.*中国).*",
		"FilterOpenAI":   "^(?!.*🇨🇳)(?!.*🇷🇺)(?!.*🇭🇰)(?!.*(中国|CN|China|俄罗斯|bRussia|RU|HongKong|HK)).*$",
	}

	// 测速地址
	speedTestURL = "http://www.gstatic.com/generate_204"
	// 手动选择图标
	selectedGroupIcon = "https://raw.githubusercontent.com/pompurin404/mihomo-party/master/resources/icon.png"
	// 自动选择图标
	autoSelectedGroupIcon = "https://fastly.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Auto.png"
	// OpenAI图标
	openAIIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/ChatGPT.png"
	// Github图标
	githubIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/GitHub.png"
	// Google图标
	googleIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/Google_Search.png"
	// 苹果图标
	appleIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/Apple_1.png"
	// 电报图标
	telegramIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/Telegram.png"
	// Netflix图标
	netflixIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/Netflix.png"
	// Disney+图标
	disneyPlusIcon = "https://fastly.jsdelivr.net/gh/Koolson/Qure@master/IconSet/Color/Disney+.png"
	// Youtube图标
	youtubeIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/YouTube.png"
	// Tiktok图标
	tiktokIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/TikTok.png"
	// Spotify图标
	spotifyIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/Spotify.png"
	// Steam 图标
	steamIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/Steam.png"
	// Gamer图标
	gamerIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/Game.png"
	// Microsoft图标
	microsoftIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/Microsoft.png"
	// GlobalMedia图标
	globalMediaIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/ForeignMedia.png"
	// 广告拦截图标
	adBlockIcon = "https://raw.githubusercontent.com/Koolson/Qure/master/IconSet/Color/AdBlack.png"
	other       = `
port: 7890
socks-port: 7891
mixed-port: 7893
redir-port: 7892
tproxy-port: 7895

unified-delay: true
geodata-mode: true
geodata-loader: standard
geo-auto-update: true
geo-update-interval: 24
tcp-concurrent: true
find-process-mode: strict
global-client-fingerprint: chrome

allow-lan: true
mode: rule
log-level: info
ipv6: true
udp: true

external-controller: 0.0.0.0:9090
# external-ui: ui
# external-ui-url: 'https://github.com/MetaCubeX/metacubexd/archive/refs/heads/gh-pages.zip'

geox-url:
  # geoip: 'https://raw.githubusercontent.com/Loyalsoldier/v2ray-rules-dat/release/geoip.dat'
  # geosite: 'https://raw.githubusercontent.com/Loyalsoldier/v2ray-rules-dat/release/geosite.dat'
  # mmdb: 'https://gitlab.com/Masaiki/GeoIP2-CN/-/raw/release/Country.mmdb'
  # asn: 'https://gitlab.com/Loon0x00/loon_data/-/raw/main/geo/GeoLite2-ASN.mmdb'
  geoip: "https://ghfast.top/https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geoip.dat"
  geosite: "https://ghfast.top/https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/geosite.dat"
  mmdb: "https://ghfast.top/https://github.com/MetaCubeX/meta-rules-dat/releases/download/latest/country.mmdb"
  asn: "https://ghfast.top//https://github.com/xishang0128/geoip/releases/download/latest/GeoLite2-ASN.mmdb"


profile:
  store-selected: true
  store-fake-ip: true

sniffer:
  enable: true
  force-dns-mapping: true
  parse-pure-ip: true
  override-destination: true
  sniff:
    HTTP:
      ports: [80, 8080-8880]
      override-destination: true
    TLS:
      ports: [443, 8443]
    QUIC:
      ports: [443, 8443]
  force-domain:
    - +.v2ex.com

  skip-domain:
    - Mijia Cloud

tun:
  enable: true
  stack: system
  dns-hijack:
    - any:53
  auto-route: true
  auto-detect-interface: true

dns:
  enable: true
  listen: 0.0.0.0:1053
  ipv6: true
  enhanced-mode: fake-ip
  fake-ip-range: 28.0.0.1/8
  fake-ip-filter:
    - "*"
    - +.lan
  default-nameserver:
    - 223.5.5.5
    - 119.29.29.29
  nameserver:
    - https://223.5.5.5/dns-query#h3=true
    - https://223.6.6.6/dns-query#h3=true

`
	rules = `
# 锚点 - 规则参数 [每天更新一次订阅规则，更新规则时使用直连策略]
rule-anchor:
  ip: &ip {type: http, interval: 86400, behavior: ipcidr, format: text}
  domain: &domain {type: http, interval: 86400, behavior: domain, format: yaml, proxy: DIRECT}
RuleSet: &RuleSet {type: http, behavior: classical, interval: 86400, format: yaml, proxy: DIRECT}
rule-providers:
  GlobalMedia:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/GlobalMedia/GlobalMedia_Classical.yaml'
    path: './RuleSet/GlobalMedia.yaml'

  ChinaMedia:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/ChinaMedia/ChinaMedia.yaml'
    path: './RuleSet/ChinaMedia.yaml'

  Netflix:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Netflix/Netflix_Classical_No_Resolve.yaml'
    path: './RuleSet/Netflix.yaml'

  Disney+:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Disney/Disney_No_Resolve.yaml'
    path: './RuleSet/Disney+.yaml'

  YouTube:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/YouTube/YouTube_No_Resolve.yaml'
    path: './RuleSet/YouTube.yaml'
          
  google_domain:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Google/Google_No_Resolve.yaml'
    path: './ruleset/google_domain.yaml'

  github_domain:
    <<: *domain
    url: 'https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/meta/geo/geosite/github.yaml'
    path: './ruleset/github_domain.yaml'

  Apple:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Apple/Apple_Classical_No_Resolve.yaml'
    path: './RuleSet/Apple.yaml'

  Microsoft:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Microsoft/Microsoft_No_Resolve.yaml'
    path: './RuleSet/Microsoft.yaml'
    
  telegramcidr:
    type: http
    behavior: ipcidr
    url: "https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/telegramcidr.txt"
    path: ./ruleset/telegramcidr.yaml
    interval: 86400

  cncidr:
    type: http
    behavior: ipcidr
    url: "https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/cncidr.txt"
    path: ./ruleset/cncidr.yaml
    interval: 86400
    
  Nintendo:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Nintendo/Nintendo.yaml'
    path: './RuleSet/Nintendo.yaml'
    
  PlayStation:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/PlayStation/PlayStation.yaml'
    path: './RuleSet/PlayStation.yaml'
    
  Epic:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Epic/Epic.yaml'
    path: './RuleSet/Epic.yaml'
    
  Xbox:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Xbox/Xbox.yaml'
    path: './RuleSet/Xbox.yaml'

  SteamDomain:
    <<: *domain
    url: 'https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/refs/heads/meta/geo/geosite/steam.yaml'
    path: './ruleset/steam_domain.yaml'

  SteamDomesticDomain:
    <<: *domain
    url: 'https://raw.githubusercontent.com/MetaCubeX/meta-rules-dat/refs/heads/meta/geo/geosite/steam%40cn.yaml'
    path: './ruleset/Steam_domestic_domain.yaml'

  TikTok:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/TikTok/TikTok_No_Resolve.yaml'
    path: './RuleSet/TikTok.yaml'

  Spotify:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Spotify/Spotify.yaml'
    path: './RuleSet/Spotify.yaml'

  OpenAI:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/OpenAI/OpenAI_No_Resolve.yaml'
    path: './RuleSet/OpenAI.yaml'
    
  Proxy:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Proxy/Proxy_Classical_No_Resolve.yaml'
    path: './RuleSet/Proxy.yaml'

  China:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/refs/heads/master/rule/Clash/ChinaMax/ChinaMax_Classical.yaml'
    path: './RuleSet/China.yaml'

  LAN:
    <<: *RuleSet
    url: 'https://raw.githubusercontent.com/blackmatrix7/ios_rule_script/master/rule/Clash/Lan/Lan.yaml'
    path: './RuleSet/LAN.yaml'
    
  Ad:
    type: http
    behavior: domain
    url: "https://cdn.jsdelivr.net/gh/Loyalsoldier/clash-rules@release/reject.txt"
    path: ./ruleset/ad.yaml
    interval: 86400

  adblock:
    type: http
    behavior: domain
    format: yaml
    url: https://raw.githubusercontent.com/REIJI007/AdBlock_Rule_For_Clash/main/adblock_reject.yaml
    path: ./ruleset/adblock_reject.yaml
    interval: 86400
    
# 分流规则指向
rules:
 - RULE-SET,Ad,广告拦截
 - RULE-SET,adblock,广告拦截
 - RULE-SET,github_domain,GitHub
 - RULE-SET,Apple,Apple
 - RULE-SET,Spotify,Spotify
 - RULE-SET,TikTok,TikTok
 - RULE-SET,Netflix,Netflix
 - RULE-SET,Disney+,Disney+
 - RULE-SET,google_domain,Google
 - RULE-SET,YouTube,YouTube
 - RULE-SET,telegramcidr,Telegram
 - RULE-SET,ChinaMedia,DIRECT
 - RULE-SET,GlobalMedia,GlobalMedia
 - RULE-SET,SteamDomain,Steam
 - RULE-SET,SteamDomesticDomain,Gamer
 - RULE-SET,Nintendo,Gamer
 - RULE-SET,PlayStation,Gamer
 - RULE-SET,Epic,Gamer
 - RULE-SET,Xbox,Gamer
 - RULE-SET,OpenAI,OpenAI
 - RULE-SET,Microsoft,Microsoft
 - RULE-SET,Proxy,Proxy
 - GEOSITE,CN,DIRECT
 - RULE-SET,China,DIRECT
 - RULE-SET,cncidr,DIRECT
 - RULE-SET,LAN,DIRECT
 - GEOIP,CN,DIRECT
 - MATCH,Proxy`
)

func main() {
	http.HandleFunc("/generate", handler2)
	log.Println("Listening on :8081")
	http.ListenAndServe("0.0.0.0:8081", nil)
}

func handler2(w http.ResponseWriter, r *http.Request) {
	// 获取参数
	url := r.URL.Query().Get("url")
	if url == "" {
		url = "https://ohayoo-pm.hf.space/api/v1/subscribe" // 默认值
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		token = "ouiw0lw7h3gx9fzm" // 默认值
	}
	target := r.URL.Query().Get("target")
	if target == "" {
		target = "clash"
	}
	list := r.URL.Query().Get("list")
	if list == "" {
		list = "true"
	}
	country_fallbackString := r.URL.Query().Get("country_fallback")
	country_fallback, _ := strconv.ParseBool(country_fallbackString)
	if country_fallbackString == "" {
		country_fallback = false
	}
	// 拼接上游订阅地址
	url = fmt.Sprintf("%s?token=%s&target=%s&list=%s", url, token, target, list)
	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to fetch subscription", 500)
		return
	}
	defer resp.Body.Close()

	data, _ := io.ReadAll(resp.Body)
	var config ClashConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		http.Error(w, "Failed to parse YAML", 500)
		return
	}
	log.Println("获取上游配置成功")
	// 删除上游的proxy-group
	config.ProxyGroups = []ProxyGroup{}

	// 对 proxies 按照名称排序
	sort.Slice(config.Proxies, func(i, j int) bool {
		nameI := fmt.Sprint(config.Proxies[i]["name"])
		nameJ := fmt.Sprint(config.Proxies[j]["name"])
		return nameI < nameJ
	})

	// 按国家分类
	countryGroups := make(map[string][]map[string]interface{})
	// 清洗所有节点的 name 字段，去除 [wzdnzd/aggregator] 及其前后空格
	reAggregator := regexp.MustCompile(`\s*[\[wzdnzd/aggregator\]|关注https://github.com/wzdnzd不迷路]\s*`)
	for _, proxy := range config.Proxies {
		m := proxy
		if m["name"] != nil {
			nameStr := fmt.Sprint(m["name"])
			nameStr = reAggregator.ReplaceAllString(nameStr, " ")
			nameStr = strings.Join(strings.Fields(nameStr), " ") // 合并多余空格
			m["name"] = decodeUnicodeEmojis(nameStr)

			name := fmt.Sprint(m["name"])
			// 用第一个数字前的内容作为分组名
			groupName := name
			for i, r := range name {
				if r >= '0' && r <= '9' {
					groupName = strings.TrimSpace(name[:i])
					break
				}
			}

			if strings.Contains(groupName, "\U0001F514") {
				groupName = "\U0001F514 关注"
			}

			countryGroups[groupName] = append(countryGroups[groupName], m)
		}
	}

	// 生成国家排序分类 sortedCountryGroup
	var sortedCountryGroup []string
	var sortedAllCountryGroup []string
	for groupName := range countryGroups {
		if groupName != "🇨🇳 中国" {
			sortedCountryGroup = append(sortedCountryGroup, groupName)
		}
		sortedAllCountryGroup = append(sortedAllCountryGroup, groupName)
	}
	sort.Strings(sortedCountryGroup)
	sort.Strings(sortedAllCountryGroup)

	// 当前获取的所有代理节点
	var allProxyNames []string
	for _, groupName := range sortedAllCountryGroup {
		for _, proxy := range countryGroups[groupName] {
			if proxy["name"] != nil {
				if name, ok := proxy["name"].(string); ok {
					allProxyNames = append(allProxyNames, name)
				}
			}
		}
	}

	doc := &yaml.Node{
		Kind:    yaml.DocumentNode,
		Content: []*yaml.Node{{Kind: yaml.MappingNode}}, // 根是一个映射
	}
	root := doc.Content[0]

	for key, val := range entries {
		// Key 节点
		k := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: key,
			Tag:   "!!str",
		}

		// Value 节点，带锚点
		v := &yaml.Node{
			Kind:   yaml.ScalarNode,
			Value:  val,
			Tag:    "!!str",
			Anchor: key, // 设置锚点
		}

		root.Content = append(root.Content, k, v)
	}
	// root.Content = append(root.Content, FilterNotChina)

	// 手动选择
	selectedGroup := ProxyGroup{Name: "Proxy", Type: "select", Proxies: append([]string{"自动选择"}, sortedCountryGroup...), Icon: selectedGroupIcon}
	config.ProxyGroups = append(config.ProxyGroups, selectedGroup)

	// 自动选择
	autoSelectedGroup := ProxyGroup{Name: "自动选择", Type: "url-test", URL: speedTestURL, IncludeAll: true, Filter: "*FilterNotChina", Icon: autoSelectedGroupIcon}
	config.ProxyGroups = append(config.ProxyGroups, autoSelectedGroup)

	// OpenAI自动选择
	autoOpenAISelectedGroup := ProxyGroup{Name: "OpenAI自动选择", Type: "url-test", URL: speedTestURL, IncludeAll: true, Filter: "*FilterOpenAI", Icon: autoSelectedGroupIcon}
	config.ProxyGroups = append(config.ProxyGroups, autoOpenAISelectedGroup)

	// OpenAI
	openAISelectedGroup := ProxyGroup{Name: "OpenAI", Type: "select", IncludeAll: true, Filter: "*FilterOpenAI", Proxies: []string{"Proxy", "OpenAI自动选择"}, Icon: openAIIcon}
	config.ProxyGroups = append(config.ProxyGroups, openAISelectedGroup)

	// Google
	googleSelectedGroup := ProxyGroup{Name: "Google", Type: "select", Proxies: append([]string{"Proxy"}, sortedCountryGroup...), Icon: googleIcon}
	config.ProxyGroups = append(config.ProxyGroups, googleSelectedGroup)

	// Github
	githubSelectedGroup := ProxyGroup{Name: "GitHub", Type: "select", Proxies: append([]string{"DIRECT", "Proxy"}, sortedCountryGroup...), Icon: githubIcon}
	config.ProxyGroups = append(config.ProxyGroups, githubSelectedGroup)

	// Apple
	appleSelectedGroup := ProxyGroup{Name: "Apple", Type: "select", Proxies: append([]string{"DIRECT", "Proxy"}, sortedCountryGroup...), Icon: appleIcon}
	config.ProxyGroups = append(config.ProxyGroups, appleSelectedGroup)

	// Telegram
	telegramSelectedGroup := ProxyGroup{Name: "Telegram", Type: "select", Proxies: append([]string{"Proxy"}, sortedCountryGroup...), Icon: telegramIcon}
	config.ProxyGroups = append(config.ProxyGroups, telegramSelectedGroup)

	// Netflix
	netflixSelectedGroup := ProxyGroup{Name: "Netflix", Type: "select", Proxies: append([]string{"Proxy"}, sortedCountryGroup...), Icon: netflixIcon}
	config.ProxyGroups = append(config.ProxyGroups, netflixSelectedGroup)

	// Disney+
	disneyPlusSelectedGroup := ProxyGroup{Name: "Disney+", Type: "select", Proxies: append([]string{"Proxy"}, sortedCountryGroup...), Icon: disneyPlusIcon}
	config.ProxyGroups = append(config.ProxyGroups, disneyPlusSelectedGroup)

	// Youtube
	youtubeSelectedGroup := ProxyGroup{Name: "YouTube", Type: "select", Proxies: append([]string{"Proxy"}, sortedCountryGroup...), Icon: youtubeIcon}
	config.ProxyGroups = append(config.ProxyGroups, youtubeSelectedGroup)

	// Tiktok
	tiktokSelectedGroup := ProxyGroup{Name: "TikTok", Type: "select", Proxies: append([]string{"DIRECT", "Proxy"}, sortedCountryGroup...), Icon: tiktokIcon}
	config.ProxyGroups = append(config.ProxyGroups, tiktokSelectedGroup)

	// Spotify
	spotifySelectedGroup := ProxyGroup{Name: "Spotify", Type: "select", Proxies: append([]string{"Proxy"}, sortedCountryGroup...), Icon: spotifyIcon}
	config.ProxyGroups = append(config.ProxyGroups, spotifySelectedGroup)

	// Steam
	steamSelectedGroup := ProxyGroup{Name: "Steam", Type: "select", Proxies: append([]string{"DIRECT", "Proxy"}, sortedCountryGroup...), Icon: steamIcon}
	config.ProxyGroups = append(config.ProxyGroups, steamSelectedGroup)

	// Gamer
	gamerSelectedGroup := ProxyGroup{Name: "Gamer", Type: "select", Filter: "*FilterGame", Proxies: []string{"DIRECT", "Proxy"}, Icon: gamerIcon}
	config.ProxyGroups = append(config.ProxyGroups, gamerSelectedGroup)

	// Microsoft
	microsoftSelectedGroup := ProxyGroup{Name: "Microsoft", Type: "select", Proxies: append([]string{"DIRECT", "Proxy"}, sortedCountryGroup...), Icon: microsoftIcon}
	config.ProxyGroups = append(config.ProxyGroups, microsoftSelectedGroup)

	// GlobalMedia
	globalMediaSelectedGroup := ProxyGroup{Name: "GlobalMedia", Type: "select", Proxies: append([]string{"Proxy"}, sortedCountryGroup...), Icon: globalMediaIcon}
	config.ProxyGroups = append(config.ProxyGroups, globalMediaSelectedGroup)

	// 广告拦截
	adblockSelectedGroup := ProxyGroup{Name: "广告拦截", Type: "select", Proxies: []string{"REJECT", "DIRECT", "Proxy"}, Icon: adBlockIcon}
	config.ProxyGroups = append(config.ProxyGroups, adblockSelectedGroup)

	for _, countryName := range sortedAllCountryGroup {
		proxies := []string{}
		for _, n := range allProxyNames {
			if strings.Contains(n, countryName) {
				proxies = append(proxies, n)
			}
			if strings.Contains(n, "\U0001F514") {
				proxies = append(proxies, n)
			}
		}
		var group ProxyGroup
		if !country_fallback {
			group = ProxyGroup{Name: countryName, Type: "url-test", URL: speedTestURL, Hidden: true, Proxies: proxies}
		} else {
			group = ProxyGroup{Name: countryName, Type: "fallback", Interval: 10, Lazy: true, URL: speedTestURL, Timeout: 2000, DisableUDP: false, MaxFailedTimes: 3, Hidden: true, Proxies: proxies}

		}
		config.ProxyGroups = append(config.ProxyGroups, group)
	}

	config2, err := StructToNode(config)
	if err != nil {
		http.Error(w, "Failed to marshal newConfig", 500)
		return
	}

	root.Content = append(root.Content, config2.Content...)

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	err = enc.Encode(root)
	if err != nil {
		// panic(err)
		http.Error(w, err.Error(), 500)
		return
	}

	// fmt.Println(buf.String())

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	bytes, err := nodeToBytes(root, 2)
	if err != nil {
		http.Error(w, "Failed to marshal newConfig", 500)
		return
	}

	w.Write([]byte(other))
	responseString := decodeUnicodeEmojis(string(bytes))
	responseString += rules
	w.Write([]byte(responseString))
	fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "生成clash 配置成功")

}

// 添加一个辅助函数将 []string 转换为 []interface{}
func convertToInterfaceSlice(slice []string) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}

// 修改创建分组配置的辅助函数
// func createGroupConfig(config GroupConfig) map[string]interface{} {
// 	// 将结构体转换为map
// 	m := make(map[string]interface{})
// 	bytes, _ := yaml.Marshal(config)
// 	yaml.Unmarshal(bytes, &m)
// 	return m
// }

// 将 \U000xxxxx\U000xxxxx 转换为真实 emoji（如 🇨🇳）
func decodeUnicodeEmojis(input string) string {
	// 正则匹配两个 \U 开头的 UTF-32 unicode（大多数国旗 emoji 是两个）
	re := regexp.MustCompile(`\\U[0-9A-Fa-f]{8}\\U[0-9A-Fa-f]{8}`)
	return re.ReplaceAllStringFunc(input, func(m string) string {
		parts := strings.Split(m, `\U`)
		if len(parts) != 3 {
			return m
		}
		code1, err1 := strconv.ParseInt(parts[1], 16, 32)
		code2, err2 := strconv.ParseInt(parts[2], 16, 32)
		if err1 != nil || err2 != nil {
			return m
		}
		r1 := rune(code1)
		r2 := rune(code2)

		buf := make([]byte, 0, 8)
		buf = append(buf, string(r1)...)
		buf = append(buf, string(r2)...)

		if utf8.Valid(buf) {
			return string(buf)
		}
		return m
	})
}

// 递归处理字符串和数字
func toYAMLValue(v interface{}) string {
	switch vv := v.(type) {
	case string:
		return strconv.Quote(vv)
	case bool:
		return fmt.Sprintf("%v", vv)
	case int, int64, float64:
		return fmt.Sprintf("%v", vv)
	case map[string]interface{}:
		var items []string
		for k, val := range vv {
			items = append(items, k+": "+toYAMLValue(val))
		}
		return "{" + strings.Join(items, ", ") + "}"
	case []string:
		var items []string
		for _, val := range vv {
			items = append(items, strconv.Quote(val))
		}
		return "[" + strings.Join(items, ", ") + "]"
	case []interface{}:
		var items []string
		for _, val := range vv {
			str := toYAMLValue(val)
			// 如果是字符串类型且不是以引号开始，添加引号
			if _, err := strconv.ParseFloat(str, 64); err != nil && !strings.HasPrefix(str, "\"") {
				str = strconv.Quote(str)
			}
			items = append(items, str)
		}
		return "[" + strings.Join(items, ", ") + "]"
	default:
		return fmt.Sprintf("%v", vv)
	}
}

func nodeToBytes(node *yaml.Node, indent int) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indent)
	if err := enc.Encode(node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func WalkYAML(node *yaml.Node, depth int) {
	// indent := strings.Repeat("  ", depth)
	fieldOrder := []string{"name", "server", "type", "url", "include-all", "filter", "timeout", "password", "cipher"}

	// 建立字段优先级 map
	fieldPriority := map[string]int{}
	for i, key := range fieldOrder {
		fieldPriority[key] = i
	}

	switch node.Kind {
	case yaml.DocumentNode:
		// fmt.Printf("%sDocument:\n", indent)
		for _, n := range node.Content {
			WalkYAML(n, depth+1)
		}
	case yaml.MappingNode:
		// fmt.Printf("%sMapping:\n", indent)

		// 提取所有 key-value 对
		entries := make([][2]*yaml.Node, 0)
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			entries = append(entries, [2]*yaml.Node{keyNode, valNode})
		}

		// 自定义排序逻辑：先按 fieldOrder 排序，再按 key 名字
		sort.Slice(entries, func(i, j int) bool {
			ki := entries[i][0].Value
			kj := entries[j][0].Value
			pi, iok := fieldPriority[ki]
			pj, jok := fieldPriority[kj]

			if iok && jok {
				return pi < pj
			} else if iok {
				return true
			} else if jok {
				return false
			}
			return ki < kj // 两个都不在 fieldOrder 中时按字典序排
		})

		// 重写 Content
		node.Content = []*yaml.Node{}
		for _, entry := range entries {
			keyNode := entry[0]
			valNode := entry[1]
			// fmt.Printf("%s  Key: %s, value: %s\n", indent, keyNode.Value, valNode.Value)
			keyNode.Style = yaml.FlowStyle
			if keyNode.Value == "filter" {
				filterValue := "FilterNotChina"
				if valNode.Value == "*FilterNotChina" {
					filterValue = "FilterNotChina"
				} else if valNode.Value == "*FilterOpenAI" {
					filterValue = "FilterOpenAI"
				} else if valNode.Value == "*FilterGame" {
					filterValue = "FilterGame"
				}

				// 创建引用锚点的 alias 节点
				filterAlias := &yaml.Node{
					Kind:  yaml.AliasNode,
					Value: filterValue,
					// Alias: FilterNotChina,
				}

				valNode = filterAlias
			}
			node.Content = append(node.Content, keyNode, valNode)
			WalkYAML(valNode, depth+2)
		}
	case yaml.SequenceNode:
		// fmt.Printf("%sSequence:\n", indent)
		for _, item := range node.Content {
			item.Style = yaml.FlowStyle
			WalkYAML(item, depth+1)
		}
	case yaml.ScalarNode:
		// fmt.Printf("%sScalar: %s\n", indent, node.Value)
	default:
		// fmt.Printf("%sUnknown kind: %d\n", indent, node.Kind)
	}
}

func StructToNode(v interface{}) (*yaml.Node, error) {
	// 1. 先 marshal 成 []byte
	data, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}

	// 2. 然后 unmarshal 到 *yaml.Node
	var node yaml.Node
	err = yaml.Unmarshal(data, &node)
	if err != nil {
		return nil, err
	}
	WalkYAML(&node, 0)
	// node.Kind == DocumentNode，node.Content[0] 是实际根节点
	return node.Content[0], nil
}

func arrayToNode(v []interface{}) (*yaml.Node, error) {
	node := yaml.Node{Kind: yaml.SequenceNode}

	for i := 0; i < len(v); i++ {
		n, err := StructToNode(v[i])
		if err == nil {
			n.Style = yaml.FlowStyle
			node.Content = append(node.Content, n)
		} else {
			return nil, err
		}
	}

	return &node, nil
}
