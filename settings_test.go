package main

import "testing"

//Gotta test that it adds the user to the list, as well as uncapitalizes their name. 
func TestTrust(t *testing.T) {
	s := NewSettings("settings.json")
	oldTrusted := s.TrustedUsers
	s.trustUser("NewTrustedUser")
	if len(s.TrustedUsers) != (len(oldTrusted) + 1) {
		t.Errorf("Trusted array was %d long, but should have been %d long", len(s.TrustedUsers), (len(oldTrusted) + 1))	
	}
	for i, v := range oldTrusted {
		if(v != s.TrustedUsers[i]) {
			t.Errorf("Lists are not identical before new user. Expected %s at position %d, got %s", v, i, s.TrustedUsers[i])	
		}
	}
	if s.TrustedUsers[len(s.TrustedUsers)-1] != "newtrusteduser" {
		t.Errorf("Trusted user is incorrect. Expected newtrusteduser, got %s", s.TrustedUsers[len(s.TrustedUsers)-1])
	}
}
