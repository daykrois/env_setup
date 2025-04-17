package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// 定义缓存的有效期（例如 30 天）
const cacheDuration = 30 * 24 * time.Hour

// 缓存文件路径
const cacheFile = "jdk_links_cache.json"

// 定义缓存结构
type Cache struct {
	Timestamp time.Time         `json:"timestamp"` // 缓存时间戳
	Links     map[string]string `json:"links"`     // 键是 JDK 名字，值是下载链接
}

// 获取 JDK 下载链接（带缓存）
func GetJDKLinks(url string, filterKeyword string) (map[string]string, error) {
	// 检查缓存是否有效
	cache, err := loadCache()
	if err == nil && time.Since(cache.Timestamp) < cacheDuration {
		fmt.Println("正在使用缓存结果...")
		return cache.Links, nil
	}

	// 缓存无效，重新抓取
	fmt.Println("缓存无效或不存在，正在重新抓取...")
	linksFromWeb, err := fetchJDKLinks(url, filterKeyword)
	if err != nil {
		return nil, fmt.Errorf("获取 JDK 链接失败: %w", err)
	}

	// 更新缓存
	cache = Cache{
		Timestamp: time.Now(),
		Links:     linksFromWeb,
	}
	if err := saveCache(cache); err != nil {
		log.Printf("警告：无法保存缓存：%v", err)
	}

	return linksFromWeb, nil
}

// 从网页抓取 JDK 链接
func fetchJDKLinks(url string, filterKeyword string) (map[string]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("无法访问 URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器返回错误状态码: %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析 HTML 失败: %w", err)
	}

	links := make(map[string]string)
	doc.Find("tbody").Each(func(i int, tbody *goquery.Selection) {
		tbody.Find("a").Each(func(i int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if !exists {
				return
			}
			fmt.Println(href)
			if strings.Contains(href, filterKeyword) && !strings.Contains(href, "sha256") {
				filename := path.Base(href)
				jdkName := strings.Split(filename, "_")[0] // 提取 JDK 名字
				links[jdkName] = href
			}
		})
	})

	return links, nil
}

// 加载缓存
func loadCache() (Cache, error) {
	var cache Cache

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return cache, fmt.Errorf("读取缓存文件失败: %w", err)
	}

	if err := json.Unmarshal(data, &cache); err != nil {
		return cache, fmt.Errorf("解析缓存内容失败: %w", err)
	}

	return cache, nil
}

// 保存缓存
func saveCache(cache Cache) error {
	data, err := json.Marshal(cache)
	if err != nil {
		return fmt.Errorf("序列化缓存内容失败: %w", err)
	}

	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("写入缓存文件失败: %w", err)
	}

	return nil
}
