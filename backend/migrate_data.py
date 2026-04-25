import re
import os

input_path = r'database\db_douyin.sql'
output_path = r'database\data_only.sql'

# Maps table name to list of columns
table_columns = {}

current_table = None
in_create_table = False

# Regexes
re_create_table = re.compile(r'^CREATE TABLE\s+[`"]?(\w+)[`"]?\s*\(')
re_column_def = re.compile(r'^\s*[`"]?(\w+)[`"]?\s+')
re_insert = re.compile(r'^INSERT INTO\s+[`"]?(\w+)[`"]?\s+VALUES')

print(f"Reading from {input_path}...")

with open(input_path, 'r', encoding='utf-8') as f:
    lines = f.readlines()

print("Analyzing schema...")

# Pass 1: Extract column definitions from CREATE TABLE blocks
for line in lines:
    line_stripped = line.strip()
    
    # Start of CREATE TABLE
    m_create = re_create_table.match(line_stripped)
    if m_create:
        current_table = m_create.group(1)
        table_columns[current_table] = []
        in_create_table = True
        continue
        
    # End of CREATE TABLE
    if in_create_table and (line_stripped.startswith('PRIMARY KEY') or line_stripped.startswith('KEY') or line_stripped.startswith('UNIQUE KEY') or line_stripped.startswith('CONSTRAINT') or line_stripped.startswith(')')):
        # Note: We stop collecting columns when we hit keys or constraints or closing bracket
        # But we stay in 'in_create_table' until the semicolon to avoid confusing the parser
        # Actually, for column collection, we just stop collecting.
        # But for state management, we need to know when the block ends.
        # Let's just stop collecting columns.
        pass
        
    if in_create_table and (line_stripped.endswith(';') or '; ' in line_stripped):
        in_create_table = False
        current_table = None
        continue
        
    # Column definition
    if in_create_table:
        m_col = re_column_def.match(line_stripped)
        if m_col:
            col_name = m_col.group(1)
            # Ignore special keywords just in case regex is too loose
            if col_name.upper() not in ['PRIMARY', 'KEY', 'UNIQUE', 'CONSTRAINT', 'FOREIGN', ')']:
                table_columns[current_table].append(col_name)

print(f"Found columns for {len(table_columns)} tables.")

print("Generating data extraction script...")

# List of tables in the target schema (new_db_douyin.sql)
target_tables = [
    'tb_accounts', 'tb_users', 'tb_relations', 'tb_messages', 'tb_posts', 
    'tb_goods', 'tb_videos', 'tb_music', 'tb_source', 'tb_statistics', 
    'tb_collects', 'tb_comments', 'tb_diggs', 'tb_shares', 
    'tb_auth_casbin_rule', 'tb_auth_access_tokens'
]

# Pass 2: Generate data_only.sql
with open(output_path, 'w', encoding='utf-8') as f_out:
    f_out.write("SET FOREIGN_KEY_CHECKS = 0;\n")
    
    in_create_table_block = False
    for line in lines:
        stripped = line.strip()
        
        if stripped.startswith('DROP TABLE'):
            continue
            
        if stripped.startswith('CREATE TABLE'):
            in_create_table_block = True
            continue
            
        if in_create_table_block:
            if stripped.endswith(';') or '; ' in stripped: # End of create table
                in_create_table_block = False
            continue
            
        # Now we are outside CREATE TABLE.
        if stripped.startswith('INSERT INTO'):
            m_insert = re_insert.match(stripped)
            if m_insert:
                tbl = m_insert.group(1)
                
                # Skip tables not in target schema
                if tbl not in target_tables:
                    continue
                    
                if tbl in table_columns:
                    cols = table_columns[tbl]
                    
                    # Fix known column name mismatches
                    fixed_cols = []
                    for c in cols:
                        if tbl == 'tb_goods' and c == 'isLowPrice':
                            fixed_cols.append('is_low_price')
                        else:
                            fixed_cols.append(c)
                            
                    col_str = ', '.join([f'`{c}`' for c in fixed_cols])
                    # Replace "INSERT INTO tbl VALUES" with "INSERT INTO tbl (cols) VALUES"
                    # We use line.replace to keep original indentation if any, but replace the start
                    # Regex substitution is safer for the start
                    new_line = re.sub(r'INSERT INTO\s+[`"]?' + tbl + r'[`"]?\s+VALUES', f'INSERT INTO `{tbl}` ({col_str}) VALUES', line, count=1)
                    f_out.write(new_line)
                else:
                    # Fallback
                    f_out.write(line)
        else:
            # Write other lines (LOCK, UNLOCK, comments, SET...)
            # We should probably skip LOCK/UNLOCK for tables we are skipping?
            # Ideally yes, but MySQL usually ignores LOCK/UNLOCK if table doesn't exist? 
            # No, it might error.
            # Let's simple-check for LOCK TABLES tbl WRITE
            if stripped.startswith('LOCK TABLES'):
                m_lock = re.match(r'LOCK TABLES [`"]?(\w+)[`"]?', stripped)
                if m_lock and m_lock.group(1) not in target_tables:
                    continue
            
            f_out.write(line)
            
    f_out.write("\nSET FOREIGN_KEY_CHECKS = 1;\n")

print(f"Done! Data extracted to {output_path}")
