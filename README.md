# kubegen

## 功能
- 通过模板替换生成kubernetes,格式类似shell,如$VAR,${VAR},支持简单函数处理,如:${-|file logstash.conf}
    - 变量可以从控制台输入,如 -v APP=game
    - 也可以从文件输入,如-c ./values.yaml, 可以通过-s选择某个组

## 用法
- kubegen <files> [-l filename=index] [--apply] [--expand] [-i input_dir] [-o output_dir] [--prefix value] [--suffix value] [-c value_file] [-s selector] [-v key=value]
   - files: 要处理的模板文件, -l file=index 指定需要合并的文件,index是files文件索引
   - -i input_dir: files输入文件目录
   - -o output_dir: files输出文件目录, --prefix:输出文件添加前缀, --suffix:输出文件添加后缀
   - -c value_file: 文件中配置需要指定的变量, -s selector:变量分组
   - -v key=value: kv格式需要替换的变量
   - --expand:是否将spec.template.spec.imagePullSecrets这样以点分隔的key展开
   - --apply:是否调用kubectl apply, -n 指定namespace

## 例子
- kubegen service.yaml deployment.yaml --expand -l deployment-tencent.yaml=1 -c values.yaml -s tencent -i ./data/game -o ./data/game_out -v APP=commgame -v ENV=alpha
- kubegen logstash.yaml -v APP=word -v IMAGE=docker.elastic.co/logstash/logstash:7.0.1 -i ./data/logstash -o ./data/logstash_out

## 参考
- kubetpl: https://github.com/shyiko/kubetpl

