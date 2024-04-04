package v2_test

import (
	"bytes"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"github.com/formancehq/stack/libs/go-libs/time"

	"github.com/formancehq/stack/libs/go-libs/auth"
	"github.com/formancehq/stack/libs/go-libs/bun/bunpaginate"

	ledger "github.com/formancehq/ledger/internal"
	v2 "github.com/formancehq/ledger/internal/api/v2"
	"github.com/formancehq/ledger/internal/opentelemetry/metrics"
	"github.com/formancehq/ledger/internal/storage/ledgerstore"
	sharedapi "github.com/formancehq/stack/libs/go-libs/api"

	"github.com/formancehq/stack/libs/go-libs/query"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetVolumes(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name              string
		queryParams       url.Values
		body              string
		expectQuery       ledgerstore.PaginatedQueryOptions[ledgerstore.PITFilterForVolumes]
		expectStatusCode  int
		expectedErrorCode string
	}
	before := time.Now()
	zero := time.Time{}

	testCases := []testCase{
		{
			name: "basic",
			expectQuery: ledgerstore.NewPaginatedQueryOptions(ledgerstore.PITFilterForVolumes{
				PITFilter: ledgerstore.PITFilter{
					PIT:&before,
					OOT:&zero,
				},
			}).
				WithPageSize(v2.DefaultPageSize),
		},
		{
			name: "using metadata",
			body: `{"$match": { "metadata[roles]": "admin" }}`,
			expectQuery: ledgerstore.NewPaginatedQueryOptions(ledgerstore.PITFilterForVolumes{
				PITFilter: ledgerstore.PITFilter{
					PIT:&before,
					OOT:&zero,
				},
			}).
				WithQueryBuilder(query.Match("metadata[roles]", "admin")).
				WithPageSize(v2.DefaultPageSize),
		},
		{
			name: "using address",
			body: `{"$match": { "address": "foo" }}`,
			expectQuery: ledgerstore.NewPaginatedQueryOptions(ledgerstore.PITFilterForVolumes{
				PITFilter: ledgerstore.PITFilter{
					PIT:&before,
					OOT:&zero,
				},
			}).
				WithQueryBuilder(query.Match("address", "foo")).
				WithPageSize(v2.DefaultPageSize),
		},
		{
			name:              "using invalid query payload",
			body:              `[]`,
			expectStatusCode:  http.StatusBadRequest,
			expectedErrorCode: v2.ErrValidation,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {

			if testCase.expectStatusCode == 0 {
				testCase.expectStatusCode = http.StatusOK
			}

			expectedCursor := bunpaginate.Cursor[ledger.VolumesWithBalanceByAssetByAccount]{
				Data: []ledger.VolumesWithBalanceByAssetByAccount{
					{
						Account : "user:1",
						Asset: "eur",
						VolumesWithBalance: ledger.VolumesWithBalance{
							Input:big.NewInt(1),
							Output: big.NewInt(1),
							Balance: big.NewInt(0),
						},

					},
				},
			}

			backend, mockLedger := newTestingBackend(t, true)
			if testCase.expectStatusCode < 300 && testCase.expectStatusCode >= 200 {
				mockLedger.EXPECT().
				GetVolumesWithBalances(gomock.Any(), ledgerstore.NewGetVolumesWithBalancesQuery(testCase.expectQuery)).
					Return(&expectedCursor, nil)
			}

			router := v2.NewRouter(backend, nil, metrics.NewNoOpRegistry(), auth.NewNoAuth())

			req := httptest.NewRequest(http.MethodGet, "/xxx/volumes?pit="+before.Format(time.RFC3339Nano), bytes.NewBufferString(testCase.body))
			rec := httptest.NewRecorder()
			params := url.Values{}
			if testCase.queryParams != nil {
				params = testCase.queryParams
			}
			params.Set("pit", before.Format(time.RFC3339Nano))
			req.URL.RawQuery = params.Encode()

			router.ServeHTTP(rec, req)

			require.Equal(t, testCase.expectStatusCode, rec.Code)
			if testCase.expectStatusCode < 300 && testCase.expectStatusCode >= 200 {
				cursor := sharedapi.DecodeCursorResponse[ledger.VolumesWithBalanceByAssetByAccount](t, rec.Body)
				require.Equal(t, expectedCursor, *cursor)
			} else {
				err := sharedapi.ErrorResponse{}
				sharedapi.Decode(t, rec.Body, &err)
				require.EqualValues(t, testCase.expectedErrorCode, err.ErrorCode)
			}
		})
	}
}