package providers

import (
	"regexp"
	"strings"
)

// NonAlphaNumRegex matches any character that is not a lowercase alphanumeric character.
var NonAlphaNumRegex = regexp.MustCompile(`[^a-z0-9]`)

// EnglishStopWordsRaw is the canonical constant list of common English stop words.
const EnglishStopWordsRaw = "a,about,above,after,again,all,am,an,and,any,are,as,at,be,because,been,before,being,below,between,both,but,by,can,could,did,do,does,doing,down,during,each,few,for,from,further,had,has,have,having,he,her,here,hers,him,his,how,i,if,in,into,is,it,its,just,me,more,most,my,no,nor,not,of,off,on,once,only,or,other,our,out,over,own,same,she,should,so,some,such,than,that,the,their,theirs,them,then,there,these,they,this,those,through,to,too,under,until,up,very,was,we,were,what,when,where,which,while,who,whom,why,will,with,you,your"

// EnglishStopWords provides O(1) membership lookup for common English stop words.
var EnglishStopWords = initStopWords(EnglishStopWordsRaw)

func initStopWords(raw string) map[string]bool {
	words := strings.Split(raw, ",")
	m := make(map[string]bool, len(words))
	for _, word := range words {
		m[word] = true
	}
	return m
}

// IsStopWord checks if a word is an English stop word.
func IsStopWord(w string) bool {
	return EnglishStopWords[strings.ToLower(strings.TrimSpace(w))]
}
