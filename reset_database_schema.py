#!/usr/bin/env python3
"""
Database schema reset script - drops all tables and creates fresh schema
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
            ssl_disabled=True
        )
        return connection
    except Exception as e:
        print(f"Failed to connect to database: {e}")
        return None

def reset_database_schema(connection):
    """Drop all existing tables and reset schema"""
    cursor = connection.cursor()

    print("=" * 60)
    print("RESETTING DATABASE SCHEMA")
    print("=" * 60)

    # Disable foreign key checks
    cursor.execute("SET FOREIGN_KEY_CHECKS = 0")

    # Get all existing tables
    cursor.execute("SHOW TABLES")
    tables = cursor.fetchall()

    if tables:
        print(f"\nDropping {len(tables)} existing tables:")
        for table in tables:
            table_name = table[0]
            print(f"  - Dropping table: {table_name}")
            cursor.execute(f"DROP TABLE IF EXISTS `{table_name}`")

    # Re-enable foreign key checks
    cursor.execute("SET FOREIGN_KEY_CHECKS = 1")

    print("\nSUCCESS: All existing tables dropped!")
    print("The Go backend will now create fresh schema on startup.")

    cursor.close()

def main():
    print("Connecting to Railway MySQL database...")
    connection = connect_to_database()

    if not connection:
        print("ERROR: Could not connect to database")
        return

    print("SUCCESS: Connected to Railway MySQL database")

    try:
        reset_database_schema(connection)
        print("\n" + "=" * 60)
        print("DATABASE RESET COMPLETE!")
        print("=" * 60)
        print("You can now start the application with fresh schema.")
        print("The Go backend will create all tables automatically.")

    except Exception as e:
        print(f"ERROR: Error during schema reset: {e}")
    finally:
        connection.close()
        print("\nDatabase connection closed")

if __name__ == "__main__":
    main()