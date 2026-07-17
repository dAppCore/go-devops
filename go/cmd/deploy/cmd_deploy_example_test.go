package deploy

import core "dappco.re/go"

func Example_getClient() {
	opts := core.NewOptions(core.Option{Key: "url", Value: "https://coolify.example.test"})
	_, r := getClient(opts)
	core.Println(r.OK)
	// Output: true
}
