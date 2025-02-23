package watcher

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/SongZihuan/https-watcher/src/config"
	"github.com/SongZihuan/https-watcher/src/logger"
	"github.com/SongZihuan/https-watcher/src/notify"
	"net/http"
	"time"
)

var errNotTLS = fmt.Errorf("no TLS connection was made")

func Run() error {
	if !config.IsReady() {
		panic("config is not ready")
	}

	now := time.Now()

MainCycle:
	for _, url := range config.GetConfig().Watcher.URLs {
		logger.Infof("开始请求 %s", url.Name)

		tlsState, err := getCertificate(url.URL)
		if err != nil {
			if errors.Is(err, errNotTLS) {
				logger.Errorf("请求 %s 出现异常：未返回TLS证书", url.Name)
				continue MainCycle
			}

			logger.Errorf("请求 %s 出现异常：%s", url.Name, err.Error())
			continue MainCycle
		}

		if len(tlsState.PeerCertificates) == 0 {
			logger.Errorf("请求 %s 出现异常：证书链为空", url.Name)
			continue MainCycle
		}

		logger.Infof("开始处理 %s", url.Name)

		if now.After(tlsState.PeerCertificates[0].NotAfter) {
			// 证书已过期
			logger.Infof("%s 已过期", url.Name)
			notify.NewRecord(url.Name, url.URL, 0)
		} else if deadline := tlsState.PeerCertificates[0].NotAfter.Sub(now); deadline <= url.DeadlineDuration {
			// 证书即将过期
			logger.Infof("%s 即将过期", url.Name)
			notify.NewRecord(url.Name, url.URL, deadline)
		} else {
			logger.Infof("%s 正常", url.Name)
		}

		logger.Infof("处理 %s 完成", url.Name)
	}

	notify.SendNotify()

	return nil
}

func getCertificate(url string) (*tls.ConnectionState, error) {
	// 创建一个自定义的Transport，这样我们可以访问TLS连接状态
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 忽略服务器证书验证
	}

	// 使用自定义的Transport创建一个HTTP客户端
	client := &http.Client{Transport: tr}

	// 发送请求
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// 从响应中获取TLS连接状态
	tlsState := resp.TLS
	if tlsState == nil {
		return nil, errNotTLS
	}

	return tlsState, nil
}
