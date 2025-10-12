#!/usr/bin/env python3
"""
Fetch user credentials from Railway MySQL database
"""
try:
    import pymysql
    USE_PYMYSQL = True
except ImportError:
    try:
        import mysql.connector
        USE_PYMYSQL = False
    except ImportError:
        print("Error: Neither pymysql nor mysql-connector-python is installed")
        print("Install one with: pip install pymysql")
        exit(1)

import os

# Database configuration from .env
DB_CONFIG = {
    'host': 'maglev.proxy.rlwy.net',
    'port': 12371,
    'database': 'railway',
    'user': 'root',
    'password': 'seYqYtFMYUVhAexDvNyyQStGtxwKpkEf'
}

def fetch_users():
    """Fetch all active users from the database."""
    try:
        # Connect to database
        if USE_PYMYSQL:
            conn = pymysql.connect(**DB_CONFIG, cursorclass=pymysql.cursors.DictCursor)
            cursor = conn.cursor()
        else:
            conn = mysql.connector.connect(**DB_CONFIG)
            cursor = conn.cursor(dictionary=True)
        
        # Query users
        query = """
        SELECT id, name, email, role, active, last_login, created_at
        FROM users
        WHERE active = 1
        ORDER BY role, created_at
        """
        
        cursor.execute(query)
        users = cursor.fetchall()
        
        if not users:
            print("No users found in database!")
            return
        
        print("\n" + "="*80)
        print("HEALTHSECURE - USER CREDENTIALS FROM DATABASE")
        print("="*80)
        print(f"\nTotal Active Users: {len(users)}\n")
        
        # Group by role
        roles = {}
        for user in users:
            role = user['role']
            if role not in roles:
                roles[role] = []
            roles[role].append(user)
        
        # Display by role
        for role, role_users in sorted(roles.items()):
            print(f"\n{role.upper()} Users:")
            print("-" * 80)
            for user in role_users:
                print(f"  ID: {user['id']}")
                print(f"  Name: {user['name']}")
                print(f"  Email: {user['email']}")
                print(f"  Last Login: {user['last_login'] or 'Never'}")
                print(f"  Created: {user['created_at']}")
                print()
        
        # Show demo credentials
        print("\n" + "="*80)
        print("DEMO LOGIN CREDENTIALS (for frontend)")
        print("="*80)
        
        doctor = next((u for u in users if u['role'] == 'doctor'), None)
        nurse = next((u for u in users if u['role'] == 'nurse'), None)
        admin = next((u for u in users if u['role'] == 'admin'), None)
        
        if doctor:
            print(f"\nDoctor: {doctor['email']} / Doctor123")
        if nurse:
            print(f"Nurse: {nurse['email']} / Nurse123")
        if admin:
            print(f"Admin: {admin['email']} / admin123")
        
        print("\n" + "="*80)
        
        # Close connection
        cursor.close()
        conn.close()
        
        return users
        
    except Exception as e:
        print(f"Database Error: {e}")
        return None

if __name__ == "__main__":
    fetch_users()
