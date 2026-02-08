#!/bin/bash

installpackage() {
    local pkgs="$@"
    while true; do
        if ! pacman -S --noconfirm --needed $pkgs; then
            echo " -> Failed to install: $pkgs"
            echo "Retrying..."
        else
            echo " -> Installed $pkgs"
            break
        fi
    done
}


installpackage php php-sqlite php-fpm apache python composer

# This is temporary hardcoded solution!!!!!
# install composer phpmailer
cd /srv/http/NextStep
composer install --no-dev --optimize-autoloader
cd -
# End composer install

systemctl enable httpd.service

# Enable proxy modules
if ! grep -q "^LoadModule proxy_module" /etc/httpd/conf/httpd.conf; then
    echo "Enabling proxy_module..."
    sed -i 's/^#LoadModule proxy_module/LoadModule proxy_module/' /etc/httpd/conf/httpd.conf
else
    echo "proxy_module already enabled"
fi

if ! grep -q "^LoadModule proxy_fcgi_module" /etc/httpd/conf/httpd.conf; then
    echo "Enabling proxy_fcgi_module..."
    sed -i 's/^#LoadModule proxy_fcgi_module/LoadModule proxy_fcgi_module/' /etc/httpd/conf/httpd.conf
else
    echo "proxy_fcgi_module already enabled"
fi

# Create php-fpm config
mkdir -p /etc/httpd/conf/extra

echo "Creating php-fpm configuration..."
cat > /etc/httpd/conf/extra/php-fpm.conf << EOF
DirectoryIndex index.php index.html
<FilesMatch \.php$>
    SetHandler "proxy:unix:/run/php-fpm/php-fpm.sock|fcgi://localhost/"
</FilesMatch>
EOF

# Include php-fpm config
if ! grep -q "Include conf/extra/php-fpm.conf" /etc/httpd/conf/httpd.conf; then
    echo "Adding php-fpm.conf include to httpd.conf..."
    echo "Include conf/extra/php-fpm.conf" >> /etc/httpd/conf/httpd.conf
else
    echo "php-fpm.conf include already present"
fi

# Add ServerName
if ! grep -q "^ServerName" /etc/httpd/conf/httpd.conf; then
    echo "Setting ServerName to localhost..."
    echo "ServerName localhost" >> /etc/httpd/conf/httpd.conf
else
    echo "ServerName already configured"
fi

systemctl enable php-fpm.service 2>/dev/null || true

# Enable SQLite extension
if grep -q "^;extension=sqlite3" /etc/php/php.ini; then
    echo "Enabling SQLite extension..."
    sed -i 's/^;extension=sqlite3/extension=sqlite3/' /etc/php/php.ini
elif grep -q "^extension=sqlite3" /etc/php/php.ini; then
    echo "SQLite extension already enabled"
else
    echo "Adding SQLite extension..."
    echo "extension=sqlite3" >> /etc/php/php.ini
fi

echo " -> Restarting services..."
systemctl restart httpd.service
systemctl restart php-fpm.service
