-- Copyright (c) 2026 godofphonk
--
-- Permission is hereby granted, free of charge, to any person obtaining a copy
-- of this software and associated documentation files (the "Software"), to deal
-- in the Software without restriction, including without limitation the rights
-- to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
-- copies of the Software, and to permit persons to whom the Software is
-- furnished to do so, subject to the following conditions:
--
-- The above copyright notice and this permission notice shall be included in
-- all copies or substantial portions of the Software.
--
-- THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
-- IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
-- FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
-- AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
-- LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
-- OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
-- SOFTWARE.

-- Migration 011: Add telegram_id field to server_source_identifiers
-- For linking accounts between Telegram bot and web application

-- Add telegram_id column (optional field)
ALTER TABLE server_source_identifiers 
ADD COLUMN IF NOT EXISTS telegram_id BIGINT;

-- Create index for fast lookups by telegram_id
CREATE INDEX IF NOT EXISTS idx_server_source_identifiers_telegram_id 
ON server_source_identifiers(telegram_id) 
WHERE telegram_id IS NOT NULL;

-- Add comment
COMMENT ON COLUMN server_source_identifiers.telegram_id IS 'Telegram user ID for linking accounts between TG bot and web application';
