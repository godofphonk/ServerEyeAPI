// Copyright (c) 2026 godofphonk
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package interfaces

import (
	"context"
	"time"

	"github.com/godofphonk/ServerEyeAPI/internal/models"
)

type AlertRepository interface {
	Create(ctx context.Context, alert *models.Alert) error
	GetByID(ctx context.Context, alertID string) (*models.Alert, error)
	GetByServerID(ctx context.Context, serverID string, limit int) ([]*models.Alert, error)
	GetActiveByServerID(ctx context.Context, serverID string) ([]*models.Alert, error)
	GetByServerIDAndType(ctx context.Context, serverID string, alertType models.AlertType) ([]*models.Alert, error)
	GetByTimeRange(ctx context.Context, serverID string, start, end time.Time) ([]*models.Alert, error)
	Update(ctx context.Context, alert *models.Alert) error
	Resolve(ctx context.Context, alertID string) error
	ResolveByServerIDAndType(ctx context.Context, serverID string, alertType models.AlertType) error
	Delete(ctx context.Context, alertID string) error
	GetStats(ctx context.Context, serverID string, duration time.Duration) (*models.AlertStats, error)
}
