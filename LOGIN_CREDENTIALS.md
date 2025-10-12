# HealthSecure - Login Credentials

## Database Connection
**Railway MySQL Database** (Updated)
- Host: `switchback.proxy.rlwy.net`
- Port: `12240`
- Database: `railway`
- User: `root`
- Password: `trfhqfnGwkmKUaNENBXhZJpzsvUUGWEv`
- Connection String: `mysql://root:trfhqfnGwkmKUaNENBXhZJpzsvUUGWEv@switchback.proxy.rlwy.net:12240/railway`

## Default User Accounts

### 👨‍⚕️ Doctor Account
- **Email**: `dr.smith@hospital.local`
- **Password**: `Doctor123`
- **Role**: Doctor
- **Name**: Dr. John Smith

### 👩‍⚕️ Nurse Account
- **Email**: `nurse.jane@hospital.local`
- **Password**: `Nurse123`
- **Role**: Nurse
- **Name**: Nurse Jane Wilson

### 👨‍💼 Admin Account
- **Email**: `admin@hospital.local`
- **Password**: `admin123`
- **Role**: Admin
- **Name**: Admin User

## Updated Configuration

### API Keys Updated
- **Ollama API Key**: `060f7af0c2ed47b69bf82a53678cfc3f.MV0OabBAxkIkmFVSMeokvP_O` ✅

### Backend Fixed
- ✅ `GetDefaultUsersByRole()` now returns correct passwords for each role
- ✅ Frontend login component updated with correct credentials

## How to Test

1. **Restart the backend** (to apply the GetDefaultUsersByRole fix):
   ```bash
   # Stop the current backend (Ctrl+C)
   # Then restart:
   cd backend/cmd/server
   go run main.go
   ```

2. **Access the frontend**:
   - URL: http://localhost:3002
   - Select role from dropdown
   - Credentials will auto-fill
   - Click "Sign in"

3. **API Endpoint Test**:
   ```bash
   # Get default users
   curl http://localhost:8081/api/auth/defaults
   
   # Test login
   curl -X POST http://localhost:8081/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"dr.smith@hospital.local","password":"Doctor123"}'
   ```

## Notes

- All accounts are **active** by default
- Passwords are hashed using bcrypt (cost: 12)
- Demo users are seeded automatically in development mode
- Role-based access control is enforced on all endpoints
