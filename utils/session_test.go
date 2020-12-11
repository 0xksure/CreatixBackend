package utils

import "testing"

func TestIsTokenValid(t *testing.T) {
	type args struct {
		tokenString string
		secret      []byte
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := IsTokenValid(tt.args.tokenString, tt.args.secret); (err != nil) != tt.wantErr {
				t.Errorf("IsTokenValid() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
