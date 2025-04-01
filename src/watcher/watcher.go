package watcher

import (
	"crypto/tls"
	"fmt"
	"github.com/SongZihuan/https-watcher/src/config"
	"github.com/SongZihuan/https-watcher/src/logger"
	"github.com/SongZihuan/https-watcher/src/notify"
	"net/http"
	"strings"
	"time"
)

var errNotTLS = fmt.Errorf("no TLS connection was made")

func Run() error {
	if !config.IsReady() {
		panic("config is not ready")
	}

	now := time.Now()

	for _, url := range config.GetConfig().Watcher.URLs {
		logger.Infof("开始请求 %s", url.Name)

		tlsState, err := getCertificateRetry(url.URL, url.Name)
		if err != nil {
			logger.Errorf("请求 %s 出现异常：%s", url.Name, err.Error())
			notify.NewErrorRecord(url.Name, url.URL, err.Error(), url.Mark)
		} else if len(tlsState.PeerCertificates) == 0 {
			logger.Errorf("请求 %s 出现异常：证书链为空", url.Name)
			notify.NewErrorRecord(url.Name, url.URL, "证书链为空", url.Mark)
		} else {
			logger.Infof("开始处理 %s 证书", url.Name)

			if now.After(tlsState.PeerCertificates[0].NotAfter) {
				// 证书已过期
				logger.Infof("%s 已过期", url.Name)
				notify.NewOutOfDateRecord(url.Name, url.URL, 0, url.Mark)
			} else if deadline := tlsState.PeerCertificates[0].NotAfter.Sub(now); deadline <= url.DeadlineDuration {
				// 证书即将过期
				logger.Infof("%s 即将过期", url.Name)
				notify.NewOutOfDateRecord(url.Name, url.URL, deadline, url.Mark)
			} else {
				logger.Infof("%s 正常", url.Name)
			}
		}

		logger.Infof("处理 %s 完成", url.Name)
	}

	notify.SendNotify()

	return nil
}

func getCertificateRetry(url string, name string) (*tls.ConnectionState, error) {
	var err1, err2, err3 error
	var tlsStats *tls.ConnectionState

	tlsStats, err1 = getCertificate(url)
	if err1 == nil {
		return tlsStats, nil
	}

	tlsStats, err2 = getCertificate(url)
	if err2 == nil {
		return tlsStats, nil
	}

	tlsStats, err3 = getCertificate(url)
	if err3 == nil {
		return tlsStats, nil
	}

	// 去除重复
	var errMap = make(map[string]bool, 3)
	errMap[err1.Error()] = true
	errMap[err2.Error()] = true
	errMap[err3.Error()] = true

	var errStrBuilder strings.Builder
	var n = 0
	for err, _ := range errMap {
		n += 1
		errStrBuilder.WriteString(fmt.Sprintf("检查 %s 错误[%d]: %s; ", name, n, err))
	}

	err := fmt.Errorf("%s", strings.TrimSpace(errStrBuilder.String()))
	return nil, err
}

func getCertificate(url string) (*tls.ConnectionState, error) {
	// 创建一个自定义的Transport，这样我们可以访问TLS连接状态
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 忽略服务器证书验证（因为我的目的只是检查获取到的证书是否到期）
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
