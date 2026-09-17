package milvus

import (
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

func createCollectionOption(config CollectionConfig) milvusclient.CreateCollectionOption {

	schema := entity.NewSchema().WithName(config.Name).WithDescription(config.Description)

	for _, fieldConfig := range config.Fields {

		field := entity.NewField().
			WithName(fieldConfig.Name).
			WithDataType(fieldConfig.DataType).
			WithIsPrimaryKey(fieldConfig.PrimaryKey).
			WithIsAutoID(fieldConfig.AutoID)

		if fieldConfig.Dimension > 0 {
			field.WithDim(fieldConfig.Dimension)
		}

		if fieldConfig.MaxLength > 0 {
			field.WithMaxLength(fieldConfig.MaxLength)
		}

		schema.WithField(field)
	}

	option := milvusclient.NewCreateCollectionOption(
		config.Name,
		schema,
	)

	if config.ShardNum > 0 {
		option.WithShardNum(config.ShardNum)
	}

	return option
}
