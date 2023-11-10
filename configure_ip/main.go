//go:build darwin || linux

package main

import (
    "fmt"
    "os"
    // "io/ioutil"
    // "regexp"
    lib "github.com/aerospike/aerospike-management-lib"
    // "github.com/aerospike/aerospike-management-lib/asconfig"
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

    // schemaMap, err := schema.NewSchemaMap()

    managementLibLogger = logr.Logger{}
    // asconfig.InitFromMap(managementLibLogger, schemaMap)
}

func setAddress(stat lib.Stats, address string) {
    s := lib.NewSyncStats(stat)
    addresses := stat.Get("addresses")
    fmt.Printf("addresses %T %v", addresses, addresses)
    if addresses != nil {
        s.Del("addresses")
    }
    s.Set("address", address)
    fmt.Printf("%T %v\n", s, s)

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

        srcFormat := asconf.AsConfig
        outFmt := asconf.AsConfig
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

        // m1 *lib.Stats
        m1 := ac.ToMap()

        // Get and set map field
        n:=m1.Get("network")
        fabric:=n.(lib.Stats).Get("fabric")
        fmt.Printf("fabric: %T %+v\n", fabric,fabric)
        setAddress(fabric.(lib.Stats), c.String("address"))

        service:=n.(lib.Stats).Get("service")
        fmt.Printf("service: %T %+v\n", service, service)
        setAddress(service.(lib.Stats), c.String("address"))


        info:=n.(lib.Stats).Get("info")
        fmt.Printf("info: %T %+v\n", info, info)
        setAddress(info.(lib.Stats), c.String("address"))

        s := lib.NewSyncStats(*m1)
        s.Set("network", n)

        ac.LoadMap(m1)
        b := ac.ToConfText()
        // fmt.Printf("b %b", string(b))
        
        var outFile *os.File
        outPath := srcPath
        outFile, err = os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
        if err!=nil {
         panic(err)
        }
        defer outFile.Close()
        outFile.Write(b)
        return nil
    }

    app.Run(os.Args)
}