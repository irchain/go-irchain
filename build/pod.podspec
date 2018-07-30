Pod::Spec.new do |spec|
  spec.name         = 'Ghuc'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/happyuc-project/happyuc-go'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS HappyUC Client'
  spec.source       = { :git => 'https://github.com/happyuc-project/happyuc-go.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Ghuc.framework'

	spec.prepare_command = <<-CMD
    curl https://ghucstore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Ghuc.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
