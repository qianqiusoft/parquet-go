package Layout

import (
	. "github.com/xitongsys/parquet-go/Common"
	. "github.com/xitongsys/parquet-go/ParquetEncoding"
	"github.com/xitongsys/parquet-go/parquet"
	"reflect"
)

//Chunk stores the ColumnChunk in parquet file
type Chunk struct {
	Pages       []*Page
	ChunkHeader *parquet.ColumnChunk
}

//Convert several pages to one chunk
func PagesToChunk(pages []*Page) *Chunk {
	ln := len(pages)
	var numValues int64 = 0
	var totalUncompressedSize int64 = 0
	var totalCompressedSize int64 = 0

	var maxVal interface{} = pages[0].MaxVal
	var minVal interface{} = pages[0].MinVal

	for i := 0; i < ln; i++ {
		if pages[i].Header.DataPageHeader != nil {
			numValues += int64(pages[i].Header.DataPageHeader.NumValues)
		} else {
			numValues += int64(pages[i].Header.DataPageHeaderV2.NumValues)
		}
		totalUncompressedSize += int64(pages[i].Header.UncompressedPageSize) + int64(len(pages[i].RawData)) - int64(pages[i].Header.CompressedPageSize)
		totalCompressedSize += int64(len(pages[i].RawData))
		maxVal = Max(maxVal, pages[i].MaxVal)
		minVal = Min(minVal, pages[i].MinVal)
	}

	chunk := new(Chunk)
	chunk.Pages = pages
	chunk.ChunkHeader = parquet.NewColumnChunk()
	metaData := parquet.NewColumnMetaData()
	metaData.Type = pages[0].DataType
	metaData.Encodings = append(metaData.Encodings, parquet.Encoding_RLE)
	metaData.Encodings = append(metaData.Encodings, parquet.Encoding_BIT_PACKED)
	metaData.Encodings = append(metaData.Encodings, parquet.Encoding_PLAIN)
	//metaData.Encodings = append(metaData.Encodings, parquet.Encoding_DELTA_BINARY_PACKED)
	metaData.Codec = pages[0].CompressType
	metaData.NumValues = numValues
	metaData.TotalCompressedSize = totalCompressedSize
	metaData.TotalUncompressedSize = totalUncompressedSize
	metaData.PathInSchema = pages[0].DataTable.Path[1:]
	metaData.Statistics = parquet.NewStatistics()

	tmpBufMax := WritePlain([]interface{}{maxVal})
	tmpBufMin := WritePlain([]interface{}{minVal})
	name := reflect.TypeOf(maxVal).Name()

	if name == "UTF8" || name == "DECIMAL" {
		tmpBufMax = tmpBufMax[4:]
		tmpBufMin = tmpBufMin[4:]
	}
	metaData.Statistics.Max = tmpBufMax
	metaData.Statistics.Min = tmpBufMin

	chunk.ChunkHeader.MetaData = metaData
	return chunk
}
