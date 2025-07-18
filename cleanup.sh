#!/usr/bin/env bash

os=$(awk -F '=' 'tolower($0) ~ /^id=/{print $2}' /etc/*-release)
if [ "$os" != "ubuntu" ]; then
    echo "$os is not supported."
    echo "For more details refer VSCode Extension for FLOGO documentation"
    return 1
fi

readonly USER_HOME_DIR=$HOME
readonly FLOGO_BASE_DIR="$USER_HOME_DIR/tibco"
readonly CSG_LICENSE_PATH="./LICENSE.txt"


function cleanup_oracle_driver() {
    echo "removing oracle driver..."
    ora_dir="${FLOGO_BASE_DIR}/oracle"
    oracle_driver_path="${ora_dir}/oracle_odbc/instantclient_21_13"
    oracle_odbc_start="[Oracle ODBC]"
    oracle_odbc_end="FileUsage=1"

    if [ -d $ora_dir ]; then
        cleanup_directory $ora_dir
    fi

    cleanup_ld_lib_path $oracle_driver_path
    cleanup_odbc_entry $oracle_odbc_start $oracle_odbc_end
    
    if [[ $(lsb_release -rs) == "24.04" ]]; then
        sudo apt purge "libaio1t64"
    else
        sudo apt purge libaio1
    fi
    
    echo "done..."
}

function cleanup_directory() {
    cleanup_dir=$1
    echo "Removing $cleanup_dir"
    sudo rm -rf ${cleanup_dir}
        if [ $? -eq 1 ]; then
        echo "error ocurred during directory cleanup..."
        return
    fi
    echo "Removed $cleanup_dir successfully..."
}

function cleanup_ld_lib_path() {

    path_to_remove=$1
    bashrc_path="$HOME/.bashrc"
    profile_path="$HOME/.profile"

    # Check if .bashrc contains LD_LIBRARY_PATH
    if ! grep -q "LD_LIBRARY_PATH" "$bashrc_path"; then
        echo "No LD_LIBRARY_PATH found in .bashrc"
        exit 0
    else
        # Extract the current LD_LIBRARY_PATH from .bashrc
        current_ld_lib_path=$(grep 'export LD_LIBRARY_PATH=' "$bashrc_path" | sed 's/export LD_LIBRARY_PATH=//g' | tr -d '"')

        # Remove the specified path
        new_ld_lib_path=$(echo "$current_ld_lib_path" | tr ':' '\n' | grep -v "^$path_to_remove$" | paste -sd ':' -)

        # Escape slashes for sed replacement
        escaped_path=$(echo "$current_ld_lib_path" | sed 's/[\/&]/\\&/g')
        escaped_new_path=$(echo "$new_ld_lib_path" | sed 's/[\/&]/\\&/g')

        if [[ -z "$escaped_new_path" ]]; then
            sed -i "s/export LD_LIBRARY_PATH=$escaped_path//g" "$bashrc_path"
        else
            # Replace the old path with the new one in .bashrc
            sed -i "s/export LD_LIBRARY_PATH=$escaped_path/export LD_LIBRARY_PATH=$escaped_new_path/g" "$bashrc_path"
        fi
    fi

    if ! grep -q "LD_LIBRARY_PATH" "$profile_path"; then
        echo "No LD_LIBRARY_PATH found in .bashrc"
        exit 0
    else
        current_ld_lib_path=$(grep 'export LD_LIBRARY_PATH=' "$profile_path" | sed 's/export LD_LIBRARY_PATH=//g' | tr -d '"')

        # Remove the specified path
        new_ld_lib_path=$(echo "$current_ld_lib_path" | tr ':' '\n' | grep -v "^$path_to_remove$" | paste -sd ':' -)

        # Escape slashes for sed replacement
        escaped_path=$(echo "$current_ld_lib_path" | sed 's/[\/&]/\\&/g')
        escaped_new_path=$(echo "$new_ld_lib_path" | sed 's/[\/&]/\\&/g')

        if [[ -z "$escaped_new_path" ]]; then
            sed -i "s/export LD_LIBRARY_PATH=$escaped_path//g" "$profile_path"
        else
            # Replace the old path with the new one in .bashrc
            sed -i "s/export LD_LIBRARY_PATH=$escaped_path/export LD_LIBRARY_PATH=$escaped_new_path/g" "$profile_path"
        fi
    fi

    echo "Updated LD_LIBRARY_PATH..."
    #grep 'export LD_LIBRARY_PATH=' "$bashrc_path"

}

function cleanup_odbc_entry() {
    # block_start="[Oracle ODBC]"
    # block_end="FileUsage=1"
    block_start=$1
    block_end=$2
    odbcinst_file="/etc/odbcinst.ini"

    sudo sed -i "/$block_start/,/$block_end/d" $odbcinst_file
}


print_row() {
    printf "| %-20s | %-15s |\n" "$1" "$2"
}
print_line() {
    printf "+----------------------+-----------------+\n"
}

function show_help() {
    echo "Help "
    echo "This script cleans up the drivers and libraries installed using the preerquisite installation for Flogo Connectors"
    echo "Execute the script with following options to cleanup connector drivers and libraries."
    echo "For more details refer VSCode Extension for FLOGO documentation"
    print_line
    printf "|     Prerequisite     |    Argument     |\n"
    print_line
    print_row "Oracle Database  " "oracle"
    # print_row "PostgreSQL  " "postgres"
    # print_row "MySQL Database" "mysql"
    # print_row "Microsoft SQLServer" "mssql"
    # print_row "IBM MQ Client" "ibmmq"
    # print_row "UnixODBC Manager" "manager"
    print_line
    echo
    echo "  e.g ./vscode-cleanup-linux.sh mysql"
    echo
    echo "for help section, "
    echo "  ./vscode-cleanup-linux.sh help"

}

function cleanup_drivers() {

    if [ $# -eq 0 ]; then
        echo "error: no arguments provided, please provide appropriate arguments"
        show_help
        exit 1
    fi
    
    if [ -f "${CSG_LICENSE_PATH}" ]; then
        cat $CSG_LICENSE_PATH
    else
        echo "Failed to load license..."
    fi
    echo ""
    echo ""

    for arg in $@; do
        echo "***********************************"
        driver_type=$arg
        case $driver_type in
        # "postgres")
        #     echo "preparing to cleanup PostgreSQL ODBC drivers..."
        #     cleanup_psql_driver
        #     ;;
        "oracle")
            echo "preparing to cleanup Oracle ODBC drivers..."
            cleanup_oracle_driver
            ;;
        # "mysql")
        #     echo "preparing to cleanup MySQL ODBC drivers..."
        #     cleanup_mysql_driver
        #     ;;
        # "mssql")
        #     echo "preparing to cleanup MSSQL ODBC drivers..."
        #     cleanup_mssql_driver
        #     ;;
        # "manager")
        #     echo "preparing to cleanup UnixODBC Driver Manager"
        #     cleanup_unix_odbc
        #     ;;
        # "grpc")
        #     echo "preparing to cleanup grpc tooling"
        #     cleanup_grpc
        #     ;;
        # "ibmmq")
        #     echo "preparing to cleanup IBM MQ Client"
        #     cleanup_ibmmqclient
        #     ;;
        "help")
            show_help
            ;;
        *)
            echo "error: invalid argument "$driver_type" provided, please provide appropriate arguments"
            ;;
        esac
    done
}
