package commit

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Identity is an author or committer identity with timestamp.
type Identity struct {
	Name  string
	Email string
	When  time.Time
}

// Commit is a parsed Git commit object.
type Commit struct {
	Tree      string
	Parents   []string
	Author    Identity
	Committer Identity
	Message   string
}

// Serialize encodes the commit in Git's object format.
func (c *Commit) Serialize() []byte {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "tree %s\n", c.Tree)
	for _, parent := range c.Parents {
		fmt.Fprintf(&buf, "parent %s\n", parent)
	}

	writeIdentity(&buf, "author", c.Author)
	writeIdentity(&buf, "committer", c.Committer)
	buf.WriteByte('\n')
	buf.WriteString(c.Message)

	return buf.Bytes()
}

// Parse decodes a commit object body.
func Parse(content []byte) (*Commit, error) {
	parts := bytes.SplitN(content, []byte("\n\n"), 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid commit: missing message body")
	}

	headers := string(parts[0])
	message := string(parts[1])

	c := &Commit{}
	for _, line := range strings.Split(headers, "\n") {
		if line == "" {
			continue
		}

		switch {
		case strings.HasPrefix(line, "tree "):
			c.Tree = strings.TrimPrefix(line, "tree ")
		case strings.HasPrefix(line, "parent "):
			c.Parents = append(c.Parents, strings.TrimPrefix(line, "parent "))
		case strings.HasPrefix(line, "author "):
			id, err := parseIdentity(strings.TrimPrefix(line, "author "))
			if err != nil {
				return nil, err
			}
			c.Author = id
		case strings.HasPrefix(line, "committer "):
			id, err := parseIdentity(strings.TrimPrefix(line, "committer "))
			if err != nil {
				return nil, err
			}
			c.Committer = id
		default:
			return nil, fmt.Errorf("unknown commit header: %s", line)
		}
	}

	if c.Tree == "" {
		return nil, fmt.Errorf("invalid commit: missing tree")
	}
	if c.Author.Name == "" || c.Committer.Name == "" {
		return nil, fmt.Errorf("invalid commit: missing author or committer")
	}

	c.Message = message
	return c, nil
}

func writeIdentity(buf *bytes.Buffer, label string, id Identity) {
	fmt.Fprintf(buf, "%s %s <%s> %d %s\n",
		label,
		id.Name,
		id.Email,
		id.When.Unix(),
		formatZone(id.When),
	)
}

func parseIdentity(line string) (Identity, error) {
	lt := strings.IndexByte(line, '<')
	gt := strings.IndexByte(line, '>')
	if lt == -1 || gt == -1 || gt < lt {
		return Identity{}, fmt.Errorf("invalid identity: %s", line)
	}

	name := strings.TrimSpace(line[:lt])
	email := line[lt+1 : gt]
	rest := strings.Fields(strings.TrimSpace(line[gt+1:]))
	if len(rest) != 2 {
		return Identity{}, fmt.Errorf("invalid identity timestamp: %s", line)
	}

	unix, err := strconv.ParseInt(rest[0], 10, 64)
	if err != nil {
		return Identity{}, fmt.Errorf("invalid identity timestamp: %s", line)
	}

	when, err := parseZone(unix, rest[1])
	if err != nil {
		return Identity{}, err
	}

	return Identity{Name: name, Email: email, When: when}, nil
}

func formatZone(t time.Time) string {
	_, offset := t.Zone()
	sign := "+"
	if offset < 0 {
		sign = "-"
		offset = -offset
	}
	hours := offset / 3600
	mins := (offset % 3600) / 60
	return fmt.Sprintf("%s%02d%02d", sign, hours, mins)
}

func parseZone(unix int64, zone string) (time.Time, error) {
	if len(zone) != 5 {
		return time.Time{}, fmt.Errorf("invalid timezone: %s", zone)
	}

	sign := 1
	switch zone[0] {
	case '+':
	case '-':
		sign = -1
	default:
		return time.Time{}, fmt.Errorf("invalid timezone: %s", zone)
	}

	hours, err := strconv.Atoi(zone[1:3])
	if err != nil {
		return time.Time{}, err
	}
	mins, err := strconv.Atoi(zone[3:5])
	if err != nil {
		return time.Time{}, err
	}

	offset := sign * (hours*3600 + mins*60)
	return time.Unix(unix, 0).In(time.FixedZone(zone, offset)), nil
}
