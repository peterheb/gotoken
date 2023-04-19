// Copyright 2023 Peter Hebert. Licensed under the MIT license.

package main

import (
	"testing"

	"github.com/peterheb/gotoken/internal"
)

func TestTrieNode(t *testing.T) {
	// Test NewTrieNode
	root := newTrieNode(123)
	if root.value != 123 || root.index != -1 || len(root.children) != 0 {
		t.Errorf("NewTrieNode failed: unexpected default values")
	}

	// Test TrieNode.insert and TrieNode.lookup
	root = BuildTrie(wordList)
	for i, word := range wordList {
		index := root.lookup(word)
		if index != i {
			t.Errorf("TrieNode.lookup failed: expected index %d for word %s, got %d", i, word, index)
		}
	}

	// Test TrieNode.lookup for non-existing words
	nonExistingWords := []string{"xylophone", "yttrium", "zymurgy"}
	for _, word := range nonExistingWords {
		index := root.lookup(word)
		if index != -1 {
			t.Errorf("TrieNode.lookup failed: expected index -1 for non-existing word %s, got %d", word, index)
		}
	}

	// This trie is in internal/bpeParams_test.go
	//
	// smallWords := []string{"a", "b", "c", "aa", "ab", "abc"}
	// root2 := BuildTrie(smallWords)
	// ser2 := root2.serialize()
	// fmt.Printf("ser2: %#v\n", ser2)
	//
	// {3, 0x861, 0x362, 0x563, 0x102, 0x761, 0xe62, 0x501, 0xb63}

	// Test serialization and serialized lookup
	serialized := root.serialize()
	for i, word := range wordList {
		index := internal.TrieLookup(serialized, []byte(word))
		if index != i {
			t.Errorf("serialized lookup failed: expected index %d for word %s, got %08x", i, word, index)
		}
	}
}

// word list extracted from Moby Words II (public domain), deduplicated and
// shuffled
//
// source: https://www.gutenberg.org/files/3201/files/FREQ.TXT
var wordList = []string{
	"design", "half", "technique", "exist", "announce", "maybe", "improve",
	"pull", "Order", "principle", "job", "since", "home", "young", "discussion",
	"accept", "district", "catch", "saint", "stage", "soon", "hundred", "who",
	"certain", "anyone", "slowly", "an", "God", "occur", "higher", "governor",
	"effort", "mile", "hair", "thus", "behind", "few", "deal", "for", "moment",
	"language", "part", "dollar", "word", "five", "what", "forget", "create",
	"is", "feeling", "length", "march", "try", "station", "manager", "once",
	"man", "music", "total", "child", "department", "production", "live",
	"suddenly", "to", "follow", "require", "element", "knowledge", "car",
	"evening", "such", "latter", "army", "tell", "prepare", "will", "public",
	"feed", "fill", "equipment", "government", "material", "similar", "among",
	"influence", "policy", "ago", "discover", "finger", "price", "settle",
	"hard", "over", "come", "western", "poem", "agreement", "pressure", "say",
	"treatment", "spring", "offer", "class", "talk", "receive", "help", "square",
	"best", "then", "limit", "report", "publish", "level", "white", "plane",
	"movement", "nature", "something", "population", "purpose", "evidence",
	"select", "by", "floor", "back", "Brown", "drive", "could", "county",
	"enjoy", "hospital", "cover", "in", "matter", "glass", "captain", "wonder",
	"air", "he", "special", "feel", "recently", "range", "the", "single", "now",
	"employee", "association", "practice", "form", "respect", "without", "girl",
	"firm", "across", "big", "simply", "bed", "law", "research", "directly",
	"therefore", "watch", "already", "above", "learn", "place", "local",
	"serious", "condition", "lead", "indicate", "age", "most", "view", "believe",
	"drop", "analysis", "woman", "value", "different", "success", "manner",
	"real", "fear", "never", "put", "hot", "more", "university", "radio", "let",
	"action", "lose", "thought", "love", "opportunity", "here", "person",
	"building", "along", "sit", "room", "event", "so", "weapon", "secretary",
	"mean", "opinion", "than", "reduce", "close", "explain", "individual",
	"together", "number", "enter", "night", "relationship", "seek", "wait",
	"wall", "build", "remain", "just", "likely", "light", "eat", "record",
	"group", "writer", "nor", "ready", "freedom", "must", "new", "Congress",
	"certainly", "any", "series", "down", "various", "direct", "recent",
	"represent", "anything", "machine", "raise", "bad", "loss", "river", "club",
	"listen", "hear", "rather", "character", "visit", "technical", "present",
	"difficulty", "food", "morning", "instance", "Europe", "paper", "often",
	"sure", "period", "right", "term", "know", "effect", "sign", "shall",
	"themselves", "cent", "hand", "usually", "community", "permit", "normal",
	"oil", "measure", "attend", "hotel", "foot", "plant", "those", "death",
	"mouth", "program", "simple", "boy", "begin", "can", "future", "walk",
	"service", "about", "open", "Mrs", "front", "play", "model", "larger",
	"land", "officer", "sort", "apply", "method", "return", "another", "list",
	"we", "attitude", "enough", "old", "before", "how", "issue", "temperature",
	"these", "corner", "son", "court", "happen", "major", "husband", "because",
	"piece", "year", "center", "into", "upon", "see", "gas", "support", "letter",
	"rate", "fact", "several", "prove", "ship", "include", "they", "remember",
	"up", "act", "ball", "bar", "idea", "hold", "no", "find", "clear", "should",
	"stand", "physical", "school", "information", "personal", "operation",
	"fire", "other", "door", "test", "modern", "pay", "fine", "account", "six",
	"three", "house", "student", "throw", "military", "possibility", "give",
	"charge", "read", "think", "leave", "whether", "show", "produce", "game",
	"time", "that", "Mister", "agency", "large", "importance", "concern", "die",
	"experiment", "early", "tax", "police", "minute", "situation", "sale",
	"indeed", "reaction", "and", "election", "poet", "cost", "interest", "work",
	"factor", "buy", "voice", "point", "artist", "life", "too", "lay", "datum",
	"eye", "break", "source", "need", "labor", "oh", "worker", "start", "cut",
	"it", "attention", "reach", "complete", "experience", "you", "between",
	"out", "example", "project", "human", "particular", "much", "teacher",
	"itself", "problem", "tree", "leader", "perhaps", "speak", "TRUE", "name",
	"long", "south", "very", "turn", "black", "note", "body", "contain", "pass",
	"bank", "object", "tooth", "third", "grow", "stop", "image", "quite", "free",
	"lie", "case", "within", "ten", "also", "wear", "today", "far",
	"responsibility", "English", "history", "defense", "space", "date", "less",
	"low", "market", "private", "foreign", "difficult", "father", "interact",
	"road", "none", "statement", "win", "money", "dead", "page", "democratic",
	"director", "realize", "approach", "save", "although", "structure", "common",
	"wide", "per", "meeting", "mention", "faith", "book", "position", "reason",
	"sell", "compare", "French", "hope", "development", "specific", "near",
	"stock", "party", "pick", "may", "look", "step", "theory", "organization",
	"way", "mark", "distance", "while", "moral", "top", "week", "mind", "like",
	"basic", "town", "especially", "through", "yet", "marry", "all", "she",
	"story", "join", "write", "literature", "during", "end", "either", "result",
	"church", "establish", "a", "nothing", "dog", "keep", "England", "pattern",
	"own", "truth", "art", "trial", "Doctor", "use", "which", "economic",
	"there", "around", "short", "but", "science", "suggest", "move", "street",
	"recognize", "effective", "until", "spend", "facility", "patient", "every",
	"process", "fight", "base", "peace", "ride", "sometimes", "window", "trip",
	"horse", "describe", "city", "cell", "choose", "including", "business",
	"season", "office", "get", "sound", "fix", "really", "relation", "line",
	"study", "quality", "general", "leach", "again", "instead", "afternoon",
	"society", "parent", "remove", "authority", "bring", "college", "little",
	"world", "stay", "expect", "full", "well", "even", "power", "whole", "earth",
	"head", "alone", "kid", "generally", "understand", "ever", "maintain",
	"bill", "yes", "same", "set", "friend", "procedure", "I", "merely",
	"apparently", "go", "decide", "throughout", "on", "type", "finally", "small",
	"wish", "state", "continue", "as", "pool", "demand", "many", "social",
	"former", "cause", "Christian", "become", "audience", "member", "both",
	"rest", "away", "better", "ground", "final", "fall", "summer", "at", "fail",
	"why", "answer", "available", "direction", "natural", "always", "size",
	"company", "central", "disagree", "nearly", "brother", "dark",
	"administration", "religion", "from", "addition", "great", "day", "system",
	"myself", "unite", "bear", "west", "send", "cold", "last", "not", "assume",
	"supply", "would", "park", "arm", "least", "product", "table", "easy",
	"decision", "first", "rise", "express", "rule", "only", "unit", "change",
	"determine", "animal", "off", "meaning", "committee", "area", "president",
	"four", "detail", "might", "everything", "boat", "question", "kind", "hang",
	"color", "except", "seem", "under", "herself", "entire", "industrial", "of",
	"meet", "international", "inch", "do", "be", "trouble", "control", "farm",
	"water", "operate", "claim", "make", "institution", "if", "growth",
	"against", "plan", "chance", "train", "though", "hall", "agree", "still",
	"relate", "longer", "people", "lot", "next", "take", "have", "beautiful",
	"himself", "subject", "appear", "aid", "choice", "develop", "late", "board",
	"section", "marriage", "century", "role", "enemy", "scene", "kill", "field",
	"with", "item", "southern", "two", "achieve", "smile", "extend", "when",
	"hill", "actually", "each", "edge", "thing", "side", "press", "prevent",
	"provide", "according", "lady", "region", "force", "able", "run",
	"Occurrence", "mother", "heart", "want", "conference", "later", "greater",
	"carry", "affair", "however", "basis", "ask", "serve", "American", "country",
	"good", "commission", "wife", "national", "course", "face", "increase",
	"suffer", "almost", "toward", "performance", "religious", "obtain", "care",
	"important", "one", "function", "finish", "Decreasing", "vary", "where",
	"allow", "volume", "strength", "red", "month", "John", "citizen", "medical",
	"strong", "political", "gun", "probably", "strike", "sense", "family",
	"fund", "picture", "leg", "involve", "consider", "discuss", "figure",
	"million", "necessary", "second", "draw", "add", "contribute", "hour",
	"difference", "after", "immediately", "else", "call", "some", "shoot",
	"division", "suppose", "federal", "or", "hit", "high", "beyond", "this",
	"possible", "couple", "fiscal", "war", "amount",
}
