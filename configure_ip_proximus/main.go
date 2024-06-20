//go:build darwin || linux

package main

import (
    // "fmt"
    "os"
    lib "github.com/aerospike/aerospike-management-lib"
    "github.com/minting/asconfig/asconf"
    "github.com/go-logr/logr"
    "github.com/sirupsen/logrus"
    "github.com/urfave/cli"
    // "github.com/spf13/cobra"
)

var logger *logrus.Logger
var managementLibLogger logr.Logger

func init() {
    logger = logrus.New()

    fmt := logrus.TextFormatter{}
    fmt.FullTimestamp = true

    logger.SetFormatter(&fmt)

    managementLibLogger = logr.Logger{}
}


func setAddress(fieldStats lib.Stats, address string){
    // stat: service / interconnect /manage
    newStat := lib.NewSyncStats(fieldStats)
    ports := fieldStats.Get("ports")
    p := make(map[string]interface{})        
    portsMap,ok := ports.(map[interface{}]interface{})
    if ok {
        for k, v := range portsMap {
            k1, _ := lib.ToString(k)
            v1 := v.(map[string]interface{})
            v1["addresses"] = address
            p[k1] = v1
        }
    }
    newStat.Set("ports", p)
}

func main(){
    app := cli.NewApp()
    app.Flags = []cli.Flag{
        cli.StringFlag{
            Name:   "address",
            Usage:  "ip address",
        },
        cli.StringFlag{
            Name:   "version",
            Usage:  "aerospike version",
        },
        cli.StringFlag{
            Name:   "src_path",
            Usage:  "source aerospike.conf file path",
        },

    }
    
    app.Action = func(c *cli.Context) error {

        var err error
        srcPath := c.String("src_path")
        fdata, err := os.ReadFile(srcPath)
        if err != nil {
            panic(err)
        }

        srcFormat := asconf.YAML
        outFmt := asconf.YAML
        version := c.String("version")
        if err != nil {
            panic(err)
        }

        // ac: *asconf.asconf
        ac, err := asconf.NewAsconf(
            fdata,
            srcFormat,
            outFmt,
            version,
            logrus.New(),
            managementLibLogger,
        )
        if err != nil {
            panic(err)
        }
        address := c.String("address")

        m1 := ac.ToMap() // m1 type --  *lib.Stats

        service :=m1.Get("service")
        manage :=m1.Get("manage")
        interconnect :=m1.Get("interconnect")

        setAddress(service.(lib.Stats), address)
        setAddress(manage.(lib.Stats), address)
        setAddress(interconnect.(lib.Stats), address)

        proximus := lib.NewSyncStats(*m1)
        proximus.Set("service", service)
        proximus.Set("manage", manage)
        proximus.Set("interconnect", interconnect)

        ac.LoadMap(m1)

        text, err := ac.MarshalText()
        
        var outFile *os.File
        outPath := srcPath
        outFile, err = os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
        if err!=nil {
         panic(err)
        }
        defer outFile.Close()
        outFile.Write(text)
        return nil
    }

    app.Run(os.Args)
}