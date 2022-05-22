package apiaccount

type AccountName string

func (n AccountName) String() string { return string(n) }

func (n *AccountName) MaybeString() string {
	if n == nil {
		return n.String()
	}

	return ""
}

func (n AccountName) Empty() bool { return len(n) == 0 }

var InternalAccountName = AccountName("internal")

func (n AccountName) IsInternal() bool { return n == InternalAccountName }
