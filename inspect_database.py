#!/usr/bin/env python3
"""
Database inspection script to understand the existing schema
"""
import mysql.connector
import os
from dotenv import load_dotenv

# Load environment variables
load_dotenv('./configs/.env')

def connect_to_database():
    """Connect to the Railway MySQL database"""
    try:
        connection = mysql.connector.connect(
            host=os.getenv('DB_HOST', 'maglev.proxy.rlwy.net'),
            port=int(os.getenv('DB_PORT', 12371)),
            database=os.getenv('DB_NAME', 'railway'),
            user=os.getenv('DB_USER', 'root'),
            password=os.getenv('DB_PASSWORD'),
            ssl_disabled=True  # Since we set TLS_MODE=false
        )
        return connection
    except Exception as e:
        print(f"Failed to connect to database: {e}")
        return None

def inspect_database_schema(connection):
    """Inspect the database schema and tables"""
    cursor = connection.cursor()

    print("=" * 60)
    print("DATABASE SCHEMA INSPECTION")
    print("=" * 60)

    # Get all tables
    cursor.execute("SHOW TABLES")
    tables = cursor.fetchall()

    print(f"\nFound {len(tables)} tables:")
    for table in tables:
        print(f"  - {table[0]}")

    print("\n" + "=" * 60)
    print("TABLE STRUCTURES")
    print("=" * 60)

    # Inspect each table structure
    for table in tables:
        table_name = table[0]
        print(f"\nTable: {table_name}")
        print("-" * 40)

        cursor.execute(f"DESCRIBE {table_name}")
        columns = cursor.fetchall()

        print("Columns:")
        for col in columns:
            field, type_info, null, key, default, extra = col
            print(f"  - {field}: {type_info} {null} {key} {default} {extra}")

        # Count records
        cursor.execute(f"SELECT COUNT(*) FROM {table_name}")
        count = cursor.fetchone()[0]
        print(f"Records: {count}")

        # Show sample data if exists
        if count > 0 and count <= 10:
            cursor.execute(f"SELECT * FROM {table_name} LIMIT 5")
            sample_data = cursor.fetchall()
            print("Sample data:")
            for row in sample_data:
                print(f"  {row}")

    cursor.close()

def check_enum_columns(connection):
    """Check for enum columns that might cause issues"""
    cursor = connection.cursor()

    print("\n" + "=" * 60)
    print("ENUM COLUMN DETECTION")
    print("=" * 60)

    cursor.execute("""
        SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, COLUMN_TYPE
        FROM INFORMATION_SCHEMA.COLUMNS
        WHERE TABLE_SCHEMA = %s AND DATA_TYPE = 'enum'
    """, (os.getenv('DB_NAME', 'railway'),))

    enum_columns = cursor.fetchall()

    if enum_columns:
        print(f"\nWARNING: Found {len(enum_columns)} ENUM columns:")
        for table, column, data_type, column_type in enum_columns:
            print(f"  - {table}.{column}: {column_type}")
        print("\nNOTE: These columns need to be converted to VARCHAR for compatibility")
    else:
        print("\nSUCCESS: No ENUM columns found - schema is compatible!")

    cursor.close()

def check_users_table(connection):
    """Specifically check the users table for authentication"""
    cursor = connection.cursor()

    print("\n" + "=" * 60)
    print("USERS TABLE ANALYSIS")
    print("=" * 60)

    try:
        cursor.execute("SELECT COUNT(*) FROM users")
        user_count = cursor.fetchone()[0]
        print(f"Total users: {user_count}")

        if user_count > 0:
            cursor.execute("SELECT email, role, active, created_at FROM users LIMIT 10")
            users = cursor.fetchall()
            print("\nUser accounts:")
            for email, role, active, created_at in users:
                status = "Active" if active else "Inactive"
                print(f"  - {email} ({role}) {status} - Created: {created_at}")

    except Exception as e:
        print(f"ERROR: Error accessing users table: {e}")

    cursor.close()

def main():
    print("Connecting to HealthSecure Database...")
    connection = connect_to_database()

    if not connection:
        print("ERROR: Could not connect to database")
        return

    print("SUCCESS: Connected to Railway MySQL database")

    try:
        inspect_database_schema(connection)
        check_enum_columns(connection)
        check_users_table(connection)

        print("\n" + "=" * 60)
        print("RECOMMENDATIONS")
        print("=" * 60)
        print("SUCCESS: Database inspection complete!")
        print("NOTE: Check the enum columns above - they may need conversion")
        print("INFO: Use the enhanced migration system to handle existing tables")

    except Exception as e:
        print(f"ERROR: Error during inspection: {e}")
    finally:
        connection.close()
        print("\nDatabase connection closed")

if __name__ == "__main__":
    main()