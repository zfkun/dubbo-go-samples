/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

import (
	getty "github.com/apache/dubbo-getty"
	hessian "github.com/apache/dubbo-go-hessian2"
	gxlog "github.com/dubbogo/gost/log"
	"github.com/apache/dubbo-go/common/logger"
	_ "github.com/apache/dubbo-go/common/proxy/proxy_factory"
	"github.com/apache/dubbo-go/config"
	_ "github.com/apache/dubbo-go/protocol/dubbo"
	_ "github.com/apache/dubbo-go/registry/protocol"
	_ "github.com/apache/dubbo-go/filter/filter_impl"
	_ "github.com/apache/dubbo-go/cluster/cluster_impl"
	_ "github.com/apache/dubbo-go/cluster/loadbalance"
	_ "github.com/apache/dubbo-go/registry/zookeeper"

)

import (
	"github.com/apache/dubbo-go-samples/tls/go-client/pkg"
)

var (
	survivalTimeout int = 10e9
	userProvider = new(pkg.UserProvider)
)

func init(){
	config.SetConsumerService(userProvider)
	hessian.RegisterPOJO(&pkg.User{})
	clientKeyPath, _ := filepath.Abs("../certs/ca.key")
	caPemPath, _ := filepath.Abs("../certs/ca.pem")
	config.SetSslEnabled(true)
	config.SetClientTlsConfigBuilder(&getty.ClientTlsConfigBuilder{
		ClientPrivateKeyPath:          clientKeyPath,
		ClientTrustCertCollectionPath: caPemPath,
	})
}
// they are necessary:
// 		export CONF_CONSUMER_FILE_PATH="xxx"
// 		export APP_LOG_CONF_FILE="xxx"
func main() {
	config.Load()
	time.Sleep(3e9)

	gxlog.CInfo("\n\n\nstart to test dubbo")
	user := &pkg.User{}
	for i := 0;i < 10 ;i ++{
		err := userProvider.GetUser(context.TODO(), []interface{}{"A001"}, user)
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 5)
	}

	gxlog.CInfo("response result: %v\n", user)
	initSignal()
}
func initSignal() {
	signals := make(chan os.Signal, 1)
	// It is not possible to block SIGKILL or syscall.SIGSTOP
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGHUP,
		syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		sig := <-signals
		logger.Infof("get signal %s", sig.String())
		switch sig {
		case syscall.SIGHUP:
			// reload()
		default:
			time.AfterFunc(time.Duration(survivalTimeout), func() {
				logger.Warnf("app exit now by force...")
				os.Exit(1)
			})
			// The program exits normally or timeout forcibly exits.
			fmt.Println("app exit now...")
			return
		}
	}
}