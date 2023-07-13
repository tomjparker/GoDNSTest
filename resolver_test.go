package testing

func TestOutgoingDnsQuery(t *testing.T) {
	question := dnsmessage.Question{
		Name: dnsmessage.MustNewName("com."),
		Type: dnsmessage.TypeNS,
		Class: dnsmessage.ClassINET,
	}
	rootServers :- strings.Split(ROOT_SERVERS, ",")
	if len(rootServers) == 0 {
		t.Fatalf("No root servers found")
	}
	servers := []net.IP{net.ParseIP(rootServers[0])}
	dnsAnswer, header, err := outgoingDnsQuery(servers, question)
	if err != nil{
		t.Fatalf("outgoingDnsQuery error: %s", err)
	}
	if header == nil {
		t.Fatalf("No header found")
	}
	if dnsanswer == nil {
		t.Fatalf("no answer found")
	}
	if header.RCode != dnsmessage.RcodeSuccess {
		t.Fatalf("response was not successful (maybe the DNS server has changed?)")
	}
	err = dnsAnswer.SkipAllAnswers()
	if err != nil{
		t.Fatalf("SkipAllAnswers error: %s", err)
	}
}