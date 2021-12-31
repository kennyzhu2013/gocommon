# gocommon

Example：


    type ExampleRequest struct {
      Name string `json:"name"`
    }

    type ExampleJsonRsp struct {
      Version string `json:"version"`
      Session int    `json:"session"`
      State   int    `json:"state"`
    }

    func main() {
      // test service 11.
      prefixName := "/micro-registry/xtest"
      _ = registry.DefaultRegistry.Init(registry.Addrs("10.153.90.4:2379"), registry.Prefix(prefixName))
      clientTest := client.NewClient(client.Registry(registry.DefaultRegistry))

      // endpoint 为url, 请求为json参数.
      request := client.NewRequest("com.cmic.test", "Example", &ExampleRequest{Name: "John"})

      var eRsp ExampleJsonRsp
      err := clientTest.Call(context.Background(), request, &eRsp)
      if err != nil {
        pterm.FgRed.Printfln("Call failed:%v", err)
        return
      }

      //bs, _ := res.Read()
      //var eRsp ExampleJsonRsp
      //err = json.Unmarshal(bs, &eRsp)
      //if err != nil {
      //	pterm.FgRed.Printfln("Unmarshal failed:%v", err)
      //	return
      //}
      pterm.FgWhite.Println(eRsp)
    }
