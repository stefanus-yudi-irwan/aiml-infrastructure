package backend

import (
	"fmt"
	"strings"
)

type SQLServerMergeBuilder struct {
	Table         string
	Columns       []string
	PrimaryKeys   []string
	UpdateColumns []string
}

func (s SQLServerMergeBuilder) Build() string {

	placeholders := make([]string, len(s.Columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	sourceColumns := strings.Join(s.Columns, ", ")

	on := make([]string, len(s.PrimaryKeys))
	for i, pk := range s.PrimaryKeys {
		on[i] = fmt.Sprintf("target.%s = source.%s", pk, pk)
	}

	updates := make([]string, len(s.UpdateColumns))
	for i, col := range s.UpdateColumns {
		updates[i] = fmt.Sprintf("%s = source.%s", col, col)
	}

	inserts := strings.Join(s.Columns, ", ")
	values := make([]string, len(s.Columns))
	for i, col := range s.Columns {
		values[i] = fmt.Sprintf("source.%s", col)
	}

	return fmt.Sprintf(`
		MERGE %s AS target
		USING (
			VALUES (%s)
		) AS source (%s)
		ON %s

		WHEN MATCHED THEN
			UPDATE SET %s

		WHEN NOT MATCHED THEN
			INSERT (%s)
			VALUES (%s);`,
		s.Table,
		strings.Join(placeholders, ", "),
		sourceColumns,
		strings.Join(on, " AND "),
		strings.Join(updates, ", "),
		inserts,
		strings.Join(values, ", "),
	)
}
