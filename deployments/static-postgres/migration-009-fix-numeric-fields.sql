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

-- Migration 009: Fix numeric fields to support decimal values
-- For hardware info that can have decimal values (frequency, memory, etc.)

-- Convert integer fields to numeric to support decimal values
ALTER TABLE static_data.hardware_info 
ALTER COLUMN cpu_frequency_mhz TYPE NUMERIC(10,2) USING cpu_frequency_mhz::NUMERIC(10,2);

ALTER TABLE static_data.hardware_info 
ALTER COLUMN gpu_memory_gb TYPE NUMERIC(10,2) USING gpu_memory_gb::NUMERIC(10,2);

ALTER TABLE static_data.hardware_info 
ALTER COLUMN total_memory_gb TYPE NUMERIC(10,2) USING total_memory_gb::NUMERIC(10,2);

-- Add comments for clarity
COMMENT ON COLUMN static_data.hardware_info.cpu_frequency_mhz IS 'CPU frequency in MHz (can be decimal)';
COMMENT ON COLUMN static_data.hardware_info.gpu_memory_gb IS 'GPU memory in GB (can be decimal)';
COMMENT ON COLUMN static_data.hardware_info.total_memory_gb IS 'Total system memory in GB (can be decimal)';
