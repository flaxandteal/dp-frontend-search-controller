package data

// SearchType informs the name of the search type displayed on the website, the query retrieved from renderer and all the subtypes to pass to the logic
type SearchType struct {
	Name      string `bson:"name" json:"name"`
	QueryType string `bson:"query_type" json:"query_type"`
	SubTypes  string `bson:"sub_types" json:"sub_types"`
}

// Category contains the high level search categories and it's corresponding search types
// If search types are added or removed in the map, make sure to do the same in the defaultContentTypes variable in dp-search-query
var Category = map[string][]SearchType{
	"Publication": {Bulletin, Article, Compendium},
	"Data":        {TimeSeries, Datasets, UserRequestedData},
	"Other":       {Methodology, CorporateInformation},
}

// Bulletin - Search information specific for statistical bulletins
var Bulletin = SearchType{
	Name:      "Statistical bulletins",
	QueryType: "bulletin",
	SubTypes:  "bulletin",
}

// Article - Search information specific for articles
var Article = SearchType{
	Name:      "Article",
	QueryType: "article",
	SubTypes:  "article,article_download",
}

// Compendium - Search information specific for compendium
var Compendium = SearchType{
	Name:      "Compendium",
	QueryType: "compendia",
	SubTypes:  "compendium_landing_page,compendium_chapter",
}

// TimeSeries - Search information specific for time series
var TimeSeries = SearchType{
	Name:      "Time series",
	QueryType: "time_series",
	SubTypes:  "timeseries",
}

// Datasets - Search information specific for datasets
var Datasets = SearchType{
	Name:      "Datasets",
	QueryType: "datasets",
	SubTypes:  "dataset,dataset_landing_page,compendium_data,reference_tables,timeseries_dataset",
}

// UserRequestedData - Search information specific for user requested data
var UserRequestedData = SearchType{
	Name:      "User requested data",
	QueryType: "user_requested_data",
	SubTypes:  "static_adhoc",
}

// Methodology - Search information specific for methodologies
var Methodology = SearchType{
	Name:      "Methodology",
	QueryType: "methodology",
	SubTypes:  "static_methodology,static_methodology_download,static_qmi",
}

// CorporateInformation - Search information specific for corporate information
var CorporateInformation = SearchType{
	Name:      "Corporate Information",
	QueryType: "corporate_information",
	SubTypes:  "static_foi,static_page,static_landing_page,static_article",
}
