import os
import zipfile
import json
import xml.etree.ElementTree as ET
import argparse

def create_file(filepath):
    with open(filepath, 'w') as f:
        pass
    print(f"Created file: {filepath}")

def write_file(filepath, data):
    with open(filepath, 'a') as f:
        f.write(data + '\n')
    print(f"Data written to: {filepath}")

def read_file(filepath):
    with open(filepath, 'r') as f:
        data = f.read()
    print(f"Data: {data}")
    return data

def create_json_object(filepath, name, age):
    filepath = filepath + '.json'
    json_data = {"name": name, "age": age}
    with open(filepath, 'w') as f:
        json.dump(json_data, f, indent=2)
    print(f"JSON data saved to: {filepath}")

def create_xml_object(filepath, name, age):
    filepath = filepath + '.xml'
    root = ET.Element("Person")
    ET.SubElement(root, "Name").text = name
    ET.SubElement(root, "Age").text = age
    tree = ET.ElementTree(root)
    tree.write(filepath)
    print(f"XML data saved to: {filepath}")

def create_zip(zip_path, file_path):
    with zipfile.ZipFile(zip_path, 'w') as zf:
        zf.write(file_path, os.path.basename(file_path))
    print(f"File {file_path} added to zip: {zip_path}")

def add_to_zip(zip_path, file_path):
    with zipfile.ZipFile(zip_path, 'a') as zf:
        zf.write(file_path, os.path.basename(file_path))
    print(f"File {file_path} added to existing zip: {zip_path}")

def unzip_archive(zip_path, dest_dir):
    with zipfile.ZipFile(zip_path, 'r') as zf:
        zf.extractall(dest_dir)
    print(f"Files extracted to: {dest_dir}")

def remove_file_from_zip(zip_path, file_to_remove):
    temp_zip_path = f"{zip_path}.tmp"
    with zipfile.ZipFile(zip_path, 'r') as zf, zipfile.ZipFile(temp_zip_path, 'w') as temp_zf:
        for item in zf.infolist():
            if item.filename != file_to_remove:
                temp_zf.writestr(item, zf.read(item.filename))
    os.replace(temp_zip_path, zip_path)
    print(f"File {file_to_remove} removed from zip: {zip_path}")

def create_x_file(filepath, state, data):
    filepath = f"{filepath}.{state}"

    try:
        create_file(filepath)
    except Exception as err:
        return err

    try:
        str_data = read_file(data)
    except Exception:
        write_file(filepath, data)
    else:
        write_file(filepath, str_data)

    return None

def main():
    parser = argparse.ArgumentParser(description="File operations CLI tool")
    subparsers = parser.add_subparsers(dest="command")

    parser_create = subparsers.add_parser("createfile")
    parser_create.add_argument("filepath", type=str)

    parser_write = subparsers.add_parser("writefile")
    parser_write.add_argument("filepath", type=str)
    parser_write.add_argument("data", type=str)

    parser_read = subparsers.add_parser("readfile")
    parser_read.add_argument("filepath", type=str)

    parser_create_json_object = subparsers.add_parser("createjson")
    parser_create_json_object.add_argument("filepath", type=str)
    parser_create_json_object.add_argument("name", type=str)
    parser_create_json_object.add_argument("age", type=str)

    parser_create_xml_object = subparsers.add_parser("createxml")
    parser_create_xml_object.add_argument("filepath", type=str)
    parser_create_xml_object.add_argument("name", type=str)
    parser_create_xml_object.add_argument("age", type=str)

    parser_create_json_file = subparsers.add_parser("createjsonfile")
    parser_create_json_file.add_argument("filepath", type=str)
    parser_create_json_file.add_argument("data", type=str)

    parser_create_xml_file = subparsers.add_parser("createxmlfile")
    parser_create_xml_file.add_argument("filepath", type=str)
    parser_create_xml_file.add_argument("data", type=str)

    parser_create_zip = subparsers.add_parser("createzip")
    parser_create_zip.add_argument("zip_path", type=str)
    parser_create_zip.add_argument("file_path", type=str)

    parser_add_zip = subparsers.add_parser("addtozip")
    parser_add_zip.add_argument("zip_path", type=str)
    parser_add_zip.add_argument("file_path", type=str)

    parser_unzip = subparsers.add_parser("unzip")
    parser_unzip.add_argument("zip_path", type=str)
    parser_unzip.add_argument("dest_dir", type=str)

    parser_remove_zip = subparsers.add_parser("removefromzip")
    parser_remove_zip.add_argument("zip_path", type=str)
    parser_remove_zip.add_argument("file_to_remove", type=str)

    args = parser.parse_args()

    if args.command == "createfile":
        create_file(args.filepath)
    elif args.command == "writefile":
        write_file(args.filepath, args.data)
    elif args.command == "readfile":
        read_file(args.filepath)
    elif args.command == "createjson":
        create_json_object(args.filepath, args.name, args.age)
    elif args.command == "createxml":
        create_xml_object(args.filepath, args.name, args.age)
    elif args.command == "createzip":
        create_zip(args.zip_path, args.file_path)
    elif args.command == "addtozip":
        add_to_zip(args.zip_path, args.file_path)
    elif args.command == "unzip":
        unzip_archive(args.zip_path, args.dest_dir)
    elif args.command == "removefromzip":
        remove_file_from_zip(args.zip_path, args.file_to_remove)
    elif args.command == "createjsonfile":
        create_x_file(args.filepath, "json", args.data)
    elif args.command == "createxmlfile":
        create_x_file(args.filepath,"xml", args.data)

if __name__ == "__main__":
    main()
