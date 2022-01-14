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

//Testing untrusting the new user that was just trusted
func TestUntrustUser(t *testing.T) {
	s := NewSettings("settings.json")
	oldTrusted := s.TrustedUsers
	s.trustUser("SecondTrustedUser") //gets tested in TestTrust
	s.untrustUser("SecondTrustedUser")
	if len(s.TrustedUsers) != len(oldTrusted) {
		t.Errorf("Trusted array was %d long, but should have been %d long", len(s.TrustedUsers), len(oldTrusted))	
	}
	for i, v := range oldTrusted {
		if(v != s.TrustedUsers[i]) {
			t.Errorf("Lists are not identical after untrusting. Expected %s at position %d, got %s", v, i, s.TrustedUsers[i])	
		}
	}
	//Don't remove anyone who isn't trusted in the first place
	s.untrustUser("NotATrustedUser") 
	if len(s.TrustedUsers) != len(oldTrusted) {
		t.Errorf("Trusted array was %d long, but should have been %d long", len(s.TrustedUsers), len(oldTrusted))	
	}
	for i, v := range oldTrusted {
		if(v != s.TrustedUsers[i]) {
			t.Errorf("Lists are not identical after untrusting already untrusted user. Expected %s at position %d, got %s", v, i, s.TrustedUsers[i])	
		}
	}
}

//Gotta test that it adds the user to the list, as well as uncapitalizes their name. 
func TestIgnore(t *testing.T) {
	s := NewSettings("settings.json")
	oldIgnored := s.IgnoredUsers
	s.ignoreUser("NewIgnoredUser")
	if len(s.IgnoredUsers) != (len(oldIgnored) + 1) {
		t.Errorf("Ignored array was %d long, but should have been %d long", len(s.IgnoredUsers), (len(oldIgnored) + 1))	
	}
	for i, v := range oldIgnored {
		if(v != s.IgnoredUsers[i]) {
			t.Errorf("Lists are not identical before new user. Expected %s at position %d, got %s", v, i, s.IgnoredUsers[i])	
		}
	}
	if s.IgnoredUsers[len(s.IgnoredUsers)-1] != "newignoreduser" {
		t.Errorf("Ignored user is incorrect. Expected newignoreduser, got %s", s.IgnoredUsers[len(s.IgnoredUsers)-1])
	}
}

//Testing untrusting the new user that was just trusted
func TestUnignoreUser(t *testing.T) {
	s := NewSettings("settings.json")
	oldIgnored := s.IgnoredUsers
	s.ignoreUser("AnotherIgnoredUser") //gets tested in TestIgnore
	s.unignoreUser("AnotherIgnoredUser")
	if len(s.IgnoredUsers) != len(oldIgnored) {
		t.Errorf("Ignored array was %d long, but should have been %d long", len(s.IgnoredUsers), len(oldIgnored))	
	}
	for i, v := range oldIgnored {
		if(v != s.IgnoredUsers[i]) {
			t.Errorf("Lists are not identical after unignoring. Expected %s at position %d, got %s", v, i, s.IgnoredUsers[i])	
		}
	}
	//Don't remove anyone who isn't ignored in the first place
	s.unignoreUser("NotAnIgnoredUser") 
	if len(s.IgnoredUsers) != len(oldIgnored) {
		t.Errorf("Trusted array was %d long, but should have been %d long", len(s.IgnoredUsers), len(oldIgnored))	
	}
	for i, v := range oldIgnored {
		if(v != s.IgnoredUsers[i]) {
			t.Errorf("Lists are not identical after unignoring previously unignored user. Expected %s at position %d, got %s", v, i, s.TrustedUsers[i])	
		}
	}
}
